package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Device struct {
    ID      string
    Address string
    Secret  int
    Hash    int
}

func main() {
    // Créer 10 appareils simulés
    devices := []Device{}
    for i := 0; i < 10; i++ {
        secret := 42 + i
        device := Device{
            ID:      fmt.Sprintf("esp32_%02d", i),
            Address: fmt.Sprintf("0xSimulated%04d", i),
            Secret:  secret,
            Hash:    secret * 2,
        }
        devices = append(devices, device)
        fmt.Printf("✅ Appareil créé: %s (secret=%d, hash=%d)\n", 
                  device.ID, device.Secret, device.Hash)
    }

    var wg sync.WaitGroup
    
    // Simuler des authentifications simultanées
    for _, device := range devices {
        wg.Add(1)
        go func(d Device) {
            defer wg.Done()
            authenticateDevice(d)
        }(device)
        time.Sleep(200 * time.Millisecond)
    }
    
    wg.Wait()
    fmt.Println("\n🎉 Toutes les simulations terminées !")
}

func authenticateDevice(device Device) {
    fmt.Printf("\n🔐 [%s] Authentification ZKP...\n", device.ID)
    
    // Appel à l'API
    data := map[string]interface{}{
        "deviceAddress": device.Address,
    }
    
    jsonData, _ := json.Marshal(data)
    
    resp, err := http.Post(
        "http://localhost:8080/api/device/authenticate-zkp",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        fmt.Printf("❌ [%s] Erreur: %v\n", device.ID, err)
        return
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result["success"] == true {
        fmt.Printf("✅ [%s] Authentification RÉUSSIE !\n", device.ID)
    } else {
        fmt.Printf("❌ [%s] Authentification échouée: %v\n", 
                  device.ID, result["message"])
    }
}
