// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract IoTDeviceManager {
    
    // ============================================
    // STRUCTURES
    // ============================================
    
    struct Device {
        string publicKey;
        string metadata;
        uint256 lastSeen;
        bool isActive;
        address owner;
    }

    struct AuthRecord {
        uint256 timestamp;
        bool success;
        bytes32 proofHash;
    }
    
    struct Transaction {
        string deviceId;
        bytes data;
        string messageHash;
        uint256 sequence;
        uint256 timestamp;
        address deviceAddress;
    }

    // ============================================
    // MAPPINGS
    // ============================================
    
    mapping(address => Device) public devices;
    mapping(address => AuthRecord[]) public authHistory;
    mapping(bytes32 => bool) public usedSignatures;
    mapping(address => uint256) public authCount;
    mapping(string => Transaction[]) public deviceTransactions;
    mapping(string => uint256) public lastSequence;

    // ============================================
    // LISTE DE TOUS LES APPAREILS
    // ============================================
    
    address[] public deviceList;
    mapping(address => bool) public isDeviceListed;

    // ============================================
    // COMPTEURS GLOBAUX
    // ============================================
    
    uint256 public totalDevices;
    uint256 public totalAuthentications;
    uint256 public activeDevicesCount;

    // ============================================
    // EVENTS
    // ============================================
    
    event DeviceRegistered(address indexed deviceAddress, string publicKey);
    event DeviceAuthenticated(address indexed deviceAddress, uint256 timestamp);
    event DeviceDeactivated(address indexed deviceAddress);
    event AuthenticationRecorded(address indexed deviceAddress, uint256 timestamp, bool success);
    event TransactionRecorded(
        address indexed deviceAddress,
        string deviceId,
        uint256 sequence,
        uint256 timestamp,
        uint256 blockNumber
    );

    // ============================================
    // MODIFIERS
    // ============================================
    
    modifier onlyDeviceOwner(address deviceAddress) {
        require(msg.sender == devices[deviceAddress].owner, "Non autorise");
        _;
    }

    modifier deviceExists(address deviceAddress) {
        require(devices[deviceAddress].lastSeen != 0, "Appareil inconnu");
        _;
    }

    // ============================================
    // FONCTIONS PRINCIPALES
    // ============================================
    
    function registerDevice(
        address deviceAddress,
        string memory _publicKey,
        string memory _metadata
    ) public {
        require(devices[deviceAddress].lastSeen == 0, "Appareil deja enregistre");

        devices[deviceAddress] = Device({
            publicKey: _publicKey,
            metadata: _metadata,
            lastSeen: block.timestamp,
            isActive: true,
            owner: msg.sender
        });

        if (!isDeviceListed[deviceAddress]) {
            deviceList.push(deviceAddress);
            isDeviceListed[deviceAddress] = true;
        }

        totalDevices++;
        activeDevicesCount++;

        emit DeviceRegistered(deviceAddress, _publicKey);
    }

    function authenticateDevice(
        address deviceAddress,
        bytes32 messageHash,
        bytes memory signature
    ) public deviceExists(deviceAddress) returns (bool) {
        Device storage device = devices[deviceAddress];

        require(device.isActive, "Appareil inactif");
        require(!usedSignatures[messageHash], "Signature deja utilisee");

        address signer = recoverSigner(messageHash, signature);
        require(signer == deviceAddress, "Signature invalide");

        usedSignatures[messageHash] = true;
        device.lastSeen = block.timestamp;

        authHistory[deviceAddress].push(AuthRecord({
            timestamp: block.timestamp,
            success: true,
            proofHash: messageHash
        }));
        
        authCount[deviceAddress]++;
        totalAuthentications++;

        emit DeviceAuthenticated(deviceAddress, block.timestamp);
        emit AuthenticationRecorded(deviceAddress, block.timestamp, true);

        return true;
    }

    function recordAuthentication(
        address deviceAddress,
        bool success,
        bytes32 proofHash
    ) public deviceExists(deviceAddress) {
        
        authHistory[deviceAddress].push(AuthRecord({
            timestamp: block.timestamp,
            success: success,
            proofHash: proofHash
        }));
        
        authCount[deviceAddress]++;
        totalAuthentications++;

        emit AuthenticationRecorded(deviceAddress, block.timestamp, success);
    }

    function recordTransaction(
        address deviceAddress,
        string memory deviceId,
        bytes memory data,
        string memory messageHash,
        uint256 sequence,
        uint256 timestamp
    ) public deviceExists(deviceAddress) returns (uint256) {
        
        require(sequence > lastSequence[deviceId], "Sequence invalide");
        
        deviceTransactions[deviceId].push(Transaction({
            deviceId: deviceId,
            data: data,
            messageHash: messageHash,
            sequence: sequence,
            timestamp: timestamp,
            deviceAddress: deviceAddress
        }));
        
        lastSequence[deviceId] = sequence;
        
        emit TransactionRecorded(
            deviceAddress,
            deviceId,
            sequence,
            timestamp,
            block.number
        );
        
        return deviceTransactions[deviceId].length;
    }

    function deactivateDevice(address deviceAddress) 
        public 
        deviceExists(deviceAddress)
        onlyDeviceOwner(deviceAddress)
    {
        require(devices[deviceAddress].isActive, "Appareil deja inactif");
        
        devices[deviceAddress].isActive = false;
        activeDevicesCount--;
        
        emit DeviceDeactivated(deviceAddress);
    }

    function activateDevice(address deviceAddress) 
        public 
        deviceExists(deviceAddress)
        onlyDeviceOwner(deviceAddress)
    {
        require(!devices[deviceAddress].isActive, "Appareil deja actif");
        
        devices[deviceAddress].isActive = true;
        activeDevicesCount++;
    }

    // ============================================
    // FONCTIONS DE LECTURE
    // ============================================
    
    function getAuthHistory(address deviceAddress)
        public
        view
        deviceExists(deviceAddress)
        returns (AuthRecord[] memory)
    {
        return authHistory[deviceAddress];
    }

    function getRecentAuthHistory(address deviceAddress, uint256 limit)
        public
        view
        deviceExists(deviceAddress)
        returns (AuthRecord[] memory)
    {
        uint256 total = authHistory[deviceAddress].length;
        uint256 start = total > limit ? total - limit : 0;
        uint256 resultLength = total - start;

        AuthRecord[] memory result = new AuthRecord[](resultLength);
        for (uint256 i = 0; i < resultLength; i++) {
            result[i] = authHistory[deviceAddress][start + i];
        }
        return result;
    }

    function getGlobalStats() public view returns (
        uint256 _totalDevices,
        uint256 _totalAuthentications,
        uint256 _activeDevices,
        uint256 _inactiveDevices
    ) {
        return (
            totalDevices,
            totalAuthentications,
            activeDevicesCount,
            totalDevices - activeDevicesCount
        );
    }

    function getDeviceInfo(address deviceAddress)
        public
        view
        returns (string memory, string memory, bool, uint256)
    {
        Device storage d = devices[deviceAddress];
        return (d.publicKey, d.metadata, d.isActive, d.lastSeen);
    }

    function isDeviceRegistered(address deviceAddress)
        public
        view
        returns (bool)
    {
        return devices[deviceAddress].lastSeen != 0;
    }

    function getDeviceAuthCount(address deviceAddress)
        public
        view
        returns (uint256)
    {
        return authCount[deviceAddress];
    }

    function getAllDevices() public view returns (address[] memory) {
        return deviceList;
    }

    function getDevicesCount() public view returns (uint256) {
        return deviceList.length;
    }

    function getDevicesPaginated(uint256 start, uint256 limit) public view returns (address[] memory) {
        require(start < deviceList.length, "Start index out of bounds");
        
        uint256 end = start + limit;
        if (end > deviceList.length) {
            end = deviceList.length;
        }
        
        uint256 resultLength = end - start;
        address[] memory result = new address[](resultLength);
        
        for (uint256 i = 0; i < resultLength; i++) {
            result[i] = deviceList[start + i];
        }
        
        return result;
    }

    function getAuthHistorySize(address deviceAddress) public view returns (uint256) {
        return authHistory[deviceAddress].length;
    }

    function getTransactionCount(string memory deviceId) public view returns (uint256) {
        return deviceTransactions[deviceId].length;
    }
    
    function getTransaction(string memory deviceId, uint256 index) 
        public 
        view 
        returns (Transaction memory) 
    {
        require(index < deviceTransactions[deviceId].length, "Index invalide");
        return deviceTransactions[deviceId][index];
    }
    
    function getLatestTransaction(string memory deviceId) 
        public 
        view 
        returns (Transaction memory) 
    {
        uint256 count = deviceTransactions[deviceId].length;
        require(count > 0, "Aucune transaction");
        return deviceTransactions[deviceId][count - 1];
    }
    
    function getDeviceTransactions(string memory deviceId, uint256 start, uint256 limit) 
        public 
        view 
        returns (Transaction[] memory) 
    {
        uint256 total = deviceTransactions[deviceId].length;
        if (total == 0) {
            return new Transaction[](0);
        }
        
        uint256 end = start + limit;
        if (end > total) {
            end = total;
        }
        
        uint256 resultLength = end - start;
        Transaction[] memory result = new Transaction[](resultLength);
        
        for (uint256 i = 0; i < resultLength; i++) {
            result[i] = deviceTransactions[deviceId][start + i];
        }
        
        return result;
    }

    function recoverSigner(bytes32 _hash, bytes memory _signature)
        public
        pure
        returns (address)
    {
        require(_signature.length == 65, "Signature invalide");

        bytes32 r;
        bytes32 s;
        uint8 v;

        assembly {
            r := mload(add(_signature, 32))
            s := mload(add(_signature, 64))
            v := byte(0, mload(add(_signature, 96)))
        }

        if (v < 27) {
            v += 27;
        }

        return ecrecover(_hash, v, r, s);
    }
}
