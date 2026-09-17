#!/usr/bin/env python3
"""
detection_dev.py — Version développement CORRIGÉE v3.2
CHAIMAN IoT

Corrections v3.1 → v3.2 :
  [FIX-SCALER-5]   Si scaler.json est trouvé dans un ancêtre (models/) mais
                     pas dans model_dir, le copier automatiquement dans
                     model_dir pour les prochains lancements.

  [FIX-NORM-4]     Le z-score local utilisait le buffer complet (bon) mais
                     la normalisation était appliquée sur la séquence seule
                     au lieu du buffer complet. Correction : ref=buffer_array
                     systématiquement quand buffer_array a assez de points.

  [FIX-THRESH-3]   Le facteur NO_SCALER_THRESH_FACTOR était appliqué même
                     quand le scaler était présent. Correction : uniquement
                     si scaler is None.

  [FIX-PATH-5]     Amélioration du fallback quand aucun modèle n'est trouvé
                     pour donner un message d'erreur plus clair.
"""

from __future__ import annotations

import asyncio
import hashlib
import hmac as _hmac
import json
import logging
import math
import os
import shutil
import ssl
import threading
import time
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import aiohttp
import numpy as np

# ============================================================
# CONSTANTES
# ============================================================

_BUFFER_MAX_IDLE_SECONDS: int = int(os.getenv("BUFFER_MAX_IDLE_SECONDS", "600"))
_MAX_DEVICES: int              = int(os.getenv("MAX_DEVICES",             "10000"))
_ML_TIMEOUT_SECONDS: int       = int(os.getenv("ML_TIMEOUT_SECONDS",      "10"))
_MAX_ANALYSIS_HISTORY: int     = int(os.getenv("MAX_ANALYSIS_HISTORY",    "1000"))
_MAX_ACTION_LOG: int           = int(os.getenv("MAX_ACTION_LOG",          "200"))
_BACKEND_URL: str              = os.getenv("BACKEND_URL",                 "https://localhost:8443")

# [FIX-THRESH-3] Facteur correctif quand le scaler est absent.
_NO_SCALER_THRESH_FACTOR: float = float(os.getenv("NO_SCALER_THRESH_FACTOR", "3.0"))

# [FIX-NORM-4] Nombre minimum de points dans le buffer avant d'activer l'inférence.
_MIN_POINTS_FOR_INFERENCE: int = int(os.getenv("MIN_POINTS_FOR_INFERENCE", "50"))

# ============================================================
# DEBUG MODE
# ============================================================

_DEBUG_ML: bool = os.getenv("DEBUG_ML", "false").lower() == "true"

logging.basicConfig(
    level   = logging.DEBUG if _DEBUG_ML else logging.INFO,
    format  = "%(asctime)s [%(levelname)s] %(name)s — %(message)s",
)
logger = logging.getLogger("detection_dev")

# ============================================================
# Chargement secret HMAC
# ============================================================

def _load_ai_secret() -> str:
    secret = os.getenv("AI_BACKEND_SECRET")
    if secret:
        return secret
    default = "dev-secret-not-for-production"
    logger.warning("[DEV] AI_BACKEND_SECRET non défini — utilisation du secret de dev.")
    return default


# ============================================================
# [FIX-PATH-5] Détection du chemin du modèle et des ressources
# ============================================================

def _pick_resources_dir(model_dir: Path) -> Path:
    """
    Cherche les ressources (metadata_v4.json, scaler.json, scaler.pkl)
    dans model_dir EN PREMIER, puis dans CHAQUE ANCÊTRE jusqu'à 3 niveaux.

    [FIX-SCALER-5] Si scaler.json trouvé dans un ancêtre, le copie dans
    model_dir pour les prochains lancements.

    Cas CHAIMAN :
      model_dir = .../ml-analyzer/models/export_3datasets_v5_final
      → scaler.json absent dans model_dir
      → cherche dans .../ml-analyzer/models                  ← trouvé ici
      → copie scaler.json dans model_dir
    """
    resource_files = ["scaler.json", "metadata_v4.json", "scaler.pkl", "threshold.txt"]

    candidate = model_dir
    for _ in range(4):   # model_dir + 3 niveaux parents
        for f in resource_files:
            if (candidate / f).exists():
                logger.info(f"[PATH] Ressources trouvées dans : {candidate}")
                # [FIX-SCALER-5] Copier scaler.json dans model_dir si trouvé ailleurs
                if candidate != model_dir and (candidate / "scaler.json").exists() and not (model_dir / "scaler.json").exists():
                    try:
                        shutil.copy2(candidate / "scaler.json", model_dir / "scaler.json")
                        logger.info(f"[SCALER] Copié dans {model_dir} pour les prochains lancements")
                    except Exception as exc:
                        logger.warning(f"[SCALER] Impossible de copier : {exc}")
                return candidate
        parent = candidate.parent
        if parent == candidate:
            break
        candidate = parent

    logger.warning(
        f"[PATH] Ressources (scaler/metadata) introuvables à partir de {model_dir}.\n"
        f"       Utilisation de model_dir comme fallback."
    )
    return model_dir


def _find_model_path(env_hint: str = "") -> Tuple[Path, Path]:
    """
    Retourne (model_dir, resources_dir).
    model_dir     = répertoire contenant le .keras
    resources_dir = répertoire contenant scaler.json / metadata_v4.json
    """
    candidates: List[Path] = []

    if env_hint:
        candidates.append(Path(env_hint).expanduser())
    env_path = os.getenv("MODEL_PATH", "")
    if env_path:
        candidates.append(Path(env_path).expanduser())

    home = Path.home()
    candidates += [
        home / "iot-docker-project" / "ml-analyzer" / "models" / "export_3datasets_v5_final",
        home / "iot-docker-project" / "ml-analyzer" / "models",
        home / "iot-docker-project" / "iot-backend" / "ml-analyzer" / "models" / "export_3datasets_v5_final",
        home / "iot-project"        / "ml-analyzer" / "models" / "export_3datasets_v5_final",
        home / "iot-project"        / "ml-analyzer" / "models",
    ]

    cwd = Path.cwd()
    candidates += [
        cwd / "models" / "export_3datasets_v5_final",
        cwd / "models",
        cwd,
    ]

    def _has_model(directory: Path) -> bool:
        if not directory.exists() or not directory.is_dir():
            return False
        for ext in ["*.keras", "*.h5"]:
            if list(directory.glob(ext)):
                return True
        return (directory / "saved_model.pb").exists()

    for candidate in candidates:
        if _has_model(candidate):
            resources_dir = _pick_resources_dir(candidate)
            logger.info(f"[PATH] Modèle (.keras) : {candidate}")
            logger.info(f"[PATH] Ressources      : {resources_dir}")
            return candidate, resources_dir

        if candidate.exists() and candidate.is_dir():
            for subdir in sorted(candidate.iterdir()):
                if subdir.is_dir() and _has_model(subdir):
                    resources_dir = _pick_resources_dir(subdir)
                    logger.info(f"[PATH] Modèle (.keras) : {subdir}")
                    logger.info(f"[PATH] Ressources      : {resources_dir}")
                    return subdir, resources_dir

    fallback = candidates[0] if candidates else cwd / "models"
    logger.warning(
        f"[PATH] Aucun modèle trouvé.\n"
        f"       model_dir     : {fallback}\n"
        f"       Conseil : MODEL_PATH=<répertoire_contenant_le_.keras>"
    )
    return fallback, _pick_resources_dir(fallback)


# ============================================================
# SecureScaler
# ============================================================

@dataclass
class SecureScaler:
    mean_:           np.ndarray
    scale_:          np.ndarray
    var_:            np.ndarray
    n_samples_seen_: int

    @classmethod
    def from_json(cls, json_path: str) -> "SecureScaler":
        with open(json_path, "r", encoding="utf-8") as fh:
            data = json.load(fh)
        missing = {"mean", "scale", "var"} - data.keys()
        if missing:
            raise ValueError(f"Clés manquantes dans scaler.json : {missing}")
        return cls(
            mean_=np.array(data["mean"],  dtype=np.float64),
            scale_=np.array(data["scale"], dtype=np.float64),
            var_=np.array(data["var"],   dtype=np.float64),
            n_samples_seen_=int(data.get("n_samples_seen", 0)),
        )

    @classmethod
    def from_pickle(cls, pkl_path: str) -> "SecureScaler":
        import pickle
        logger.warning(f"[DEV] Chargement scaler pickle : {pkl_path}")
        with open(pkl_path, "rb") as fh:
            sk = pickle.load(fh)
        return cls(
            mean_=np.array(sk.mean_,   dtype=np.float64),
            scale_=np.array(sk.scale_, dtype=np.float64),
            var_=np.array(sk.var_,     dtype=np.float64),
            n_samples_seen_=int(getattr(sk, "n_samples_seen_", 0)),
        )

    def transform(self, X: np.ndarray) -> np.ndarray:
        X = np.asarray(X, dtype=np.float64)
        if X.shape[-1] != len(self.mean_):
            raise ValueError(
                f"Features : attendu {len(self.mean_)}, reçu {X.shape[-1]}"
            )
        safe_scale = np.where(np.abs(self.scale_) < 1e-6, 1.0, self.scale_)
        return (X - self.mean_) / safe_scale


# ============================================================
# MOTEUR DE DÉTECTION — v3.2
# ============================================================

class IoTNetworkAnalyzerV4Secure:

    _DEFAULT_FEATURES: List[str] = [
        "duration", "bytes_sent", "bytes_received", "packet_count",
        "bytes_per_sec", "pkts_per_sec", "bytes_per_pkt", "bytes_ratio",
        "bytes_sent_ratio", "bytes_total", "log_bytes_total",
        "pkt_size_variance", "high_volume_flag", "burst_indicator",
    ]

    def __init__(self, model_path: str = "", pubkey_path: str = "") -> None:
        self._ai_backend_secret: str = _load_ai_secret()
        self.model_path, self.resources_path = _find_model_path(model_path)

        self._load_model_keras()
        self._load_metadata()
        self._load_scaler_dev()

        # [FIX-THRESH-3] Appliquer le facteur UNIQUEMENT si scaler absent
        self._raw_threshold = self.threshold
        if self.scaler is None:
            self.threshold = self._raw_threshold * _NO_SCALER_THRESH_FACTOR
            logger.warning(
                f"[THRESH] Scaler absent — seuil ajusté : "
                f"{self._raw_threshold:.6f} × {_NO_SCALER_THRESH_FACTOR} = {self.threshold:.6f}\n"
                f"         Modifier NO_SCALER_THRESH_FACTOR si trop de faux positifs."
            )
        else:
            logger.info(f"[THRESH] Seuil avec scaler : {self.threshold:.6f}")

        if self.scaler is not None:
            logger.info(
                f"[SCALER] mean  [{self.scaler.mean_.min():.4f}, {self.scaler.mean_.max():.4f}]  "
                f"scale [{self.scaler.scale_.min():.4f}, {self.scaler.scale_.max():.4f}]  "
                f"n_features={len(self.scaler.mean_)}"
            )
        else:
            logger.warning(
                f"[SCALER] Absent — z-score local. "
                f"Inférence activée après {_MIN_POINTS_FOR_INFERENCE} points "
                f"(MIN_POINTS_FOR_INFERENCE)."
            )

        self._buf_lock     = threading.Lock()
        self._history_lock = threading.Lock()
        self._action_lock  = threading.Lock()
        self.buffers:          Dict[str, Dict[str, Any]] = {}
        self.analysis_history: List[Dict[str, Any]]      = []
        self.action_log:       List[Dict[str, Any]]       = []

        logger.info(
            f"✅ Analyseur V5 DEV v3.2 — "
            f"seuil_effectif={self.threshold:.6f}  "
            f"seq_len={self.seq_len} features={len(self.feature_cols)}  "
            f"scaler={'oui' if self.scaler else 'non (z-score local)'}"
        )

    # ---- chargement modèle ----

    def _load_model_keras(self) -> None:
        try:
            import tensorflow as tf
        except ImportError as exc:
            raise RuntimeError("pip install tensorflow") from exc

        keras_file: Optional[Path] = None
        for pattern in [
            "autoencoder_3datasets_v5*.keras",
            "autoencoder_3datasets_v5*.h5",
            "autoencoder_3datasets_v4*.keras",
            "best_model*.keras",
            "*.keras",
            "*.h5",
        ]:
            if self.model_path.exists():
                hits = sorted(
                    self.model_path.glob(pattern),
                    key=lambda p: p.stat().st_mtime,
                    reverse=True,
                )
                if hits:
                    keras_file = hits[0]
                    break

        if keras_file is None or not keras_file.exists():
            available = (
                [p.name for p in self.model_path.iterdir()]
                if self.model_path.exists() else ["<vide>"]
            )
            raise RuntimeError(
                f"Modèle introuvable dans : {self.model_path}\n"
                f"Contenu : {available}\n"
                f"Définir : MODEL_PATH=<répertoire_contenant_le_.keras>"
            )

        logger.info(f"[MODEL] Chargement : {keras_file}")
        keras_model = tf.keras.models.load_model(str(keras_file))
        logger.info(f"[MODEL] Chargé : {keras_file.name}")

        _tf = tf

        class _KerasSession:
            def get_inputs(self):
                class _I:
                    name = "input_1"
                return [_I()]

            def run(self, _, input_feed):
                data = list(input_feed.values())[0]
                inp  = _tf.constant(data, dtype=_tf.float32)
                out  = keras_model(inp, training=False)
                if isinstance(out, (list, tuple)):
                    out = out[0]
                if hasattr(out, "numpy"):
                    return [out.numpy().astype(np.float32)]
                return [np.array(out, dtype=np.float32)]

        self.session    = _KerasSession()
        self.using_onnx = False

    # ---- métadonnées ----

    def _load_metadata(self) -> None:
        for search_dir in [self.resources_path, self.model_path]:
            meta_file = search_dir / "metadata_v4.json"
            if meta_file.exists():
                with open(meta_file, "r", encoding="utf-8") as fh:
                    self.metadata = json.load(fh)
                self.threshold    = float(self.metadata.get("selected_threshold", 0.05))
                self.seq_len      = int(self.metadata.get("sequence_length", 10))
                self.feature_cols = list(self.metadata.get("feature_columns", []))
                logger.info(
                    f"[META] Chargé depuis {meta_file}\n"
                    f"       seuil={self.threshold}  seq_len={self.seq_len}  "
                    f"features={len(self.feature_cols)}"
                )
                break
        else:
            for search_dir in [self.resources_path, self.model_path]:
                thresh_file = search_dir / "threshold.txt"
                if thresh_file.exists():
                    self.threshold = float(thresh_file.read_text().strip())
                    logger.info(f"[META] Seuil depuis threshold.txt : {self.threshold}")
                    break
            else:
                self.threshold = 0.05
                logger.warning("[META] metadata_v4.json et threshold.txt absents — seuil=0.05")
            self.seq_len      = 10
            self.feature_cols = []
            self.metadata     = {}

        if not self.feature_cols:
            logger.warning("[META] features vides → liste par défaut (14 features)")
            self.feature_cols = list(self._DEFAULT_FEATURES)

    # ---- scaler ----

    def _load_scaler_dev(self) -> None:
        search_dirs = []
        seen = set()
        for d in [self.resources_path, self.model_path]:
            key = str(d)
            if key not in seen:
                seen.add(key)
                search_dirs.append(d)

        for d in search_dirs:
            j = d / "scaler.json"
            if j.exists():
                try:
                    self.scaler = SecureScaler.from_json(str(j))
                    logger.info(f"[SCALER] JSON chargé : {j}")
                    return
                except Exception as exc:
                    logger.warning(f"[SCALER] scaler.json invalide : {exc}")

        for d in search_dirs:
            for name in ["scaler.pkl", "scaler_v4.pkl", "scaler.joblib"]:
                pkl = d / name
                if pkl.exists():
                    try:
                        self.scaler = SecureScaler.from_pickle(str(pkl))
                        logger.info(f"[SCALER] Pickle chargé : {pkl}")
                        return
                    except Exception as exc:
                        logger.warning(f"[SCALER] {pkl.name} invalide : {exc}")

        self.scaler = None
        logger.warning(
            f"[SCALER] Introuvable dans : {[str(d) for d in search_dirs]}"
        )

    # ---- normalisation ----

    def _normalize(self, data: np.ndarray, buffer_data: Optional[np.ndarray] = None) -> np.ndarray:
        data = np.asarray(data, dtype=np.float64)

        if self.scaler is not None:
            try:
                return self.scaler.transform(data).astype(np.float32)
            except Exception as exc:
                logger.warning(f"[NORM] Scaler échoué ({exc}) — z-score local")

        # [FIX-NORM-4] Z-score sur le buffer complet quand disponible
        ref = (
            np.asarray(buffer_data, dtype=np.float64)
            if buffer_data is not None and len(buffer_data) >= 20
            else data
        )
        mean = np.mean(ref, axis=0, keepdims=True)
        std  = np.std(ref,  axis=0, keepdims=True)
        std  = np.where(std < 1e-6, 1.0, std)
        result = ((data - mean) / std).astype(np.float32)
        return np.clip(result, -10.0, 10.0)

    # ---- buffer ----

    def _evict_stale_buffers(self) -> int:
        now = time.monotonic()
        with self._buf_lock:
            stale = [
                a for a, b in self.buffers.items()
                if (now - b["last_seen"]) > _BUFFER_MAX_IDLE_SECONDS
            ]
            for a in stale:
                del self.buffers[a]
        if stale:
            logger.info(f"[EVICT] {len(stale)} buffer(s) inactif(s) supprimé(s)")
        return len(stale)

    def add_message(self, device_address: str, features: Dict[str, Any]) -> bool:
        with self._buf_lock:
            if len(self.buffers) >= _MAX_DEVICES and device_address not in self.buffers:
                now = time.monotonic()
                stale = [
                    a for a, b in self.buffers.items()
                    if (now - b["last_seen"]) > _BUFFER_MAX_IDLE_SECONDS
                ]
                for a in stale:
                    del self.buffers[a]
                if len(self.buffers) >= _MAX_DEVICES:
                    logger.warning(f"[BUF] MAX_DEVICES={_MAX_DEVICES} atteint")
                    return False

            if device_address not in self.buffers:
                self.buffers[device_address] = {"data": [], "last_seen": time.monotonic()}

            vec: List[float] = []
            for col in self.feature_cols:
                raw = features.get(col, 0.0)
                try:
                    v = float(raw)
                    vec.append(v if math.isfinite(v) else 0.0)
                except (TypeError, ValueError):
                    vec.append(0.0)

            buf = self.buffers[device_address]
            buf["data"].append(vec)
            buf["last_seen"] = time.monotonic()

            max_len = max(self.seq_len * 10, 200)
            if len(buf["data"]) > max_len:
                buf["data"] = buf["data"][-max_len:]

            min_required = self.seq_len if self.scaler is not None else max(self.seq_len, _MIN_POINTS_FOR_INFERENCE)
            return len(buf["data"]) >= min_required

    # ---- inférence ----

    async def analyze_device(self, device_address: str) -> Dict[str, Any]:
        with self._buf_lock:
            buf_entry   = self.buffers.get(device_address)
            snapshot    = list(buf_entry["data"][-self.seq_len:]) if buf_entry else []
            full_buffer = list(buf_entry["data"])                  if buf_entry else []

        min_required = self.seq_len if self.scaler is not None else max(self.seq_len, _MIN_POINTS_FOR_INFERENCE)

        if len(full_buffer) < min_required:
            return {
                "device_address": device_address,
                "score":          0,
                "status":         "insufficient_data",
                "action":         "none",
                "severity":       "NORMAL",
                "is_anomaly":     False,
                "message":        f"Besoin de {min_required} messages (actuel : {len(full_buffer)})",
                "timestamp":      datetime.now(tz=timezone.utc).isoformat(),
                "mse":            0.0,
                "threshold":      self.threshold,
                "model_version":  "V5 Dev v3.2",
                "f1_score":       96.45,
            }

        sequence      = np.array(snapshot,     dtype=np.float32)
        buffer_array  = np.array(full_buffer,  dtype=np.float32)
        # [FIX-NORM-4] Toujours passer buffer_array pour z-score local stable
        sequence_norm = self._normalize(sequence, buffer_array)
        sequence_norm = np.expand_dims(sequence_norm, axis=0)

        input_name    = self.session.get_inputs()[0].name
        reconstructed = self.session.run(None, {input_name: sequence_norm})[0]

        mse_raw = float(np.mean((sequence_norm - reconstructed) ** 2))
        if not math.isfinite(mse_raw) or mse_raw < 0:
            logger.warning(f"[MSE] Non fini ({mse_raw}) → 0")
            mse_raw = 0.0

        mse        = mse_raw
        is_anomaly = mse > self.threshold

        safe_threshold = max(self.threshold, 1e-8)
        score          = min(max((mse / safe_threshold) * 50.0, 0.0), 100.0)
        if not math.isfinite(score):
            score = 0.0

        if score > 90.0:   severity, action = "CRITICAL", "IMMEDIATE_REVOKE"
        elif score > 70.0: severity, action = "HIGH",     "REVOKE"
        elif score > 40.0: severity, action = "MEDIUM",   "CHALLENGE"
        elif is_anomaly:   severity, action = "LOW",      "MONITOR"
        else:              severity, action = "NORMAL",   "NONE"

        result: Dict[str, Any] = {
            "device_address": device_address,
            "score":          round(score, 2),
            "mse":            round(mse, 8),
            "threshold":      self.threshold,
            "is_anomaly":     bool(is_anomaly),
            "severity":       severity,
            "action":         action,
            "status":         "anomaly" if is_anomaly else "normal",
            "timestamp":      datetime.now(tz=timezone.utc).isoformat(),
            "model_version":  "V5 Dev v3.2",
            "f1_score":       96.45,
        }

        logger.info(
            f"[ANALYZE] {device_address[:15]}…  "
            f"MSE={mse:.6f}  seuil={self.threshold:.6f}  "
            f"score={score:.1f}  → {severity}/{action}"
        )

        with self._history_lock:
            self.analysis_history.append(result)
            if len(self.analysis_history) > _MAX_ANALYSIS_HISTORY:
                self.analysis_history = self.analysis_history[-(_MAX_ANALYSIS_HISTORY // 2):]

        if action in ("REVOKE", "IMMEDIATE_REVOKE", "CHALLENGE", "MONITOR"):
            with self._action_lock:
                self.action_log.append({
                    "id":             time.time(),
                    "device_address": device_address,
                    "action":         action,
                    "score":          round(score, 2),
                    "severity":       severity,
                    "timestamp":      datetime.now(tz=timezone.utc).isoformat(),
                })
                if len(self.action_log) > _MAX_ACTION_LOG:
                    self.action_log = self.action_log[-(_MAX_ACTION_LOG // 2):]

        if action in ("REVOKE", "IMMEDIATE_REVOKE"):
            try:
                asyncio.create_task(self._trigger_revocation(device_address, result))
            except RuntimeError:
                pass

        return result

    # ---- révocation ----

    async def _trigger_revocation(self, device_address: str, analysis: Dict[str, Any]) -> None:
        payload = {
            "deviceAddress": device_address,
            "reason": f"AI V5 DEV v3.2 — Score : {analysis['score']:.2f} ({analysis['severity']})",
            "proof": {"mse": analysis["mse"], "threshold": self.threshold, "model": "V5 Dev v3.2"},
        }
        body_bytes = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")
        mac = _hmac.new(self._ai_backend_secret.encode("utf-8"), body_bytes, hashlib.sha256)
        headers = {"Content-Type": "application/json", "X-AI-Signature": mac.hexdigest()}

        tls_verify = os.getenv("TLS_VERIFY", "true").lower()
        if tls_verify == "false":
            ssl_param: Any = False
        else:
            ssl_ctx = ssl.create_default_context()
            ca = os.getenv("BACKEND_CA_CERT")
            if ca and Path(ca).exists():
                ssl_ctx.load_verify_locations(ca)
            ssl_param = ssl_ctx

        timeout = aiohttp.ClientTimeout(total=_ML_TIMEOUT_SECONDS, connect=5)
        try:
            async with aiohttp.ClientSession(timeout=timeout) as s:
                async with s.post(f"{_BACKEND_URL}/api/device/deactivate",
                                  data=body_bytes, headers=headers, ssl=ssl_param) as resp:
                    if resp.status == 200:
                        logger.warning(f"[REVOKE] {device_address[:15]}… révoqué")
                    else:
                        # [FIX-REVOKE-1] await resp.text() d'abord, slice ensuite
                        # (évite 'coroutine' object is not subscriptable)
                        body = await resp.text()
                        logger.error(f"[REVOKE] HTTP {resp.status} : {body[:200]}")
        except Exception as exc:
            logger.error(f"[REVOKE] Erreur : {exc}")

    # ---- getters ----

    def get_device_stats(self, device_address: str) -> Dict[str, Any]:
        with self._buf_lock:
            buf = self.buffers.get(device_address)
            if buf is None:
                return {"error": "Device introuvable", "device_address": device_address}
            count    = len(buf["data"])
            raw_data = list(buf["data"])
            last_seen = buf["last_seen"]

        stats: Dict[str, Any] = {
            "device_address":     device_address,
            "messages_collected": count,
            "ready_for_analysis": count >= (self.seq_len if self.scaler else max(self.seq_len, _MIN_POINTS_FOR_INFERENCE)),
            "last_seen":          last_seen,
            "features":           self.feature_cols,
            "threshold":          self.threshold,
            "threshold_raw":      self._raw_threshold,
            "scaler_loaded":      self.scaler is not None,
        }
        if raw_data:
            arr = np.array(raw_data, dtype=np.float32)
            stats["debug"] = {
                "raw_min":  arr.min(axis=0).tolist(),
                "raw_max":  arr.max(axis=0).tolist(),
                "raw_mean": arr.mean(axis=0).tolist(),
                "raw_std":  arr.std(axis=0).tolist(),
                "n_points": count,
            }
        return stats

    def get_all_devices(self) -> List[str]:
        with self._buf_lock:
            return list(self.buffers.keys())

    def get_recent_anomalies(self, limit: int = 20) -> List[Dict[str, Any]]:
        with self._history_lock:
            return [a for a in self.analysis_history if a.get("is_anomaly")][-min(max(limit,1),500):]

    def get_action_log(self, limit: int = 20) -> List[Dict[str, Any]]:
        with self._action_lock:
            return self.action_log[-min(max(limit,1),500):]

    def get_stats(self) -> Dict[str, Any]:
        with self._history_lock:
            total = len(self.analysis_history)
            anom  = sum(1 for a in self.analysis_history if a.get("is_anomaly"))
        with self._action_lock:
            revoc = sum(1 for a in self.action_log if a.get("action") in ("REVOKE","IMMEDIATE_REVOKE"))
        with self._buf_lock:
            devs = len(self.buffers)
        return {
            "total_devices":        devs,
            "total_analyzed":       total,
            "anomalies_detected":   anom,
            "revocations":          revoc,
            "threshold":            self.threshold,
            "threshold_raw":        self._raw_threshold,
            "scaler_loaded":        self.scaler is not None,
            "no_scaler_factor":     _NO_SCALER_THRESH_FACTOR,
            "min_points_inference": _MIN_POINTS_FOR_INFERENCE,
            "f1_score":             96.45,
            "precision":            94.32,
            "recall":               98.69,
            "model_version":        "V5 Dev v3.2",
            "features":             len(self.feature_cols),
            "sequence_length":      self.seq_len,
            "using_onnx":           False,
            "model_path":           str(self.model_path),
            "resources_path":       str(self.resources_path),
        }


# ============================================================
# FastAPI
# ============================================================

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field, field_validator
import uvicorn

analyzer: Optional[IoTNetworkAnalyzerV4Secure] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global analyzer
    logger.info("=" * 60)
    logger.info("🚀 IoT Network Analyzer V5 DEV v3.2")
    logger.info("=" * 60)
    try:
        analyzer = IoTNetworkAnalyzerV4Secure()
    except Exception as exc:
        logger.critical(f"❌ Initialisation échouée : {exc}")
        raise

    async def _cleanup():
        while True:
            await asyncio.sleep(300)
            if analyzer:
                analyzer._evict_stale_buffers()

    task = asyncio.create_task(_cleanup())
    try:
        yield
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass
        logger.info("🛑 Arrêt propre")


app = FastAPI(title="IoT Network Analyzer V5 DEV v3.2", version="5.0.2-dev", lifespan=lifespan)

_ALLOWED_ORIGINS: List[str] = [
    o.strip()
    for o in os.getenv(
        "ALLOWED_ORIGINS",
        "https://localhost:3000,http://localhost:3000,http://localhost:3001,"
        "https://localhost:3001,http://localhost:8080,https://localhost:8080",
    ).split(",")
    if o.strip()
]

app.add_middleware(
    CORSMiddleware,
    allow_origins     = _ALLOWED_ORIGINS,
    allow_credentials = True,
    allow_methods     = ["GET", "POST"],
    allow_headers     = ["Content-Type", "Authorization", "X-AI-Signature"],
)


class NetworkFeatures(BaseModel):
    duration:          float = Field(0.0, ge=0, le=1e6)
    bytes_sent:        float = Field(0.0, ge=0, le=1e12)
    bytes_received:    float = Field(0.0, ge=0, le=1e12)
    packet_count:      float = Field(0.0, ge=0, le=1e6)
    bytes_per_sec:     float = Field(0.0, ge=0, le=1e9)
    pkts_per_sec:      float = Field(0.0, ge=0, le=1e6)
    bytes_per_pkt:     float = Field(0.0, ge=0, le=1e6)
    bytes_ratio:       float = Field(0.0, ge=0, le=1)
    bytes_sent_ratio:  float = Field(0.0, ge=0, le=1)
    bytes_total:       float = Field(0.0, ge=0, le=1e12)
    log_bytes_total:   float = Field(0.0, ge=0, le=50)
    pkt_size_variance: float = Field(0.0, ge=0, le=1e6)
    high_volume_flag:  float = Field(0.0, ge=0, le=1)
    burst_indicator:   float = Field(0.0, ge=0, le=100)

    @field_validator("bytes_ratio", "bytes_sent_ratio", mode="before")
    @classmethod
    def clamp_ratio(cls, v):
        return float(max(0.0, min(1.0, float(v))))

    @field_validator("log_bytes_total", mode="before")
    @classmethod
    def clamp_log(cls, v):
        return float(max(0.0, min(50.0, float(v))))


class MessageRequest(BaseModel):
    device_address: str = Field(..., min_length=42, max_length=42, pattern=r"^0x[a-fA-F0-9]{40}$")
    device_id:   Optional[str]   = None
    entity_type: Optional[str]   = None
    features:    NetworkFeatures = Field(default_factory=NetworkFeatures)

    @field_validator("device_address")
    @classmethod
    def normalize_address(cls, v):
        return v.lower()


def _get() -> IoTNetworkAnalyzerV4Secure:
    if analyzer is None:
        raise HTTPException(status_code=503, detail="Analyseur non initialisé")
    return analyzer


@app.get("/")
async def root():
    a = analyzer
    return {
        "service":        "IoT Network Analyzer V5 DEV v3.2",
        "status":         "running",
        "model_path":     str(a.model_path)     if a else "N/A",
        "resources_path": str(a.resources_path) if a else "N/A",
        "scaler_loaded":  a.scaler is not None  if a else False,
        "threshold":      a.threshold           if a else 0,
        "threshold_raw":  a._raw_threshold      if a else 0,
    }


@app.get("/health")
async def health():
    a = _get()
    return {
        "status":               "healthy",
        "model":                "V5 Dev v3.2",
        "f1_score":             96.45,
        "threshold":            a.threshold,
        "threshold_raw":        a._raw_threshold,
        "scaler_loaded":        a.scaler is not None,
        "features":             len(a.feature_cols),
        "sequence_length":      a.seq_len,
        "min_points_inference": _MIN_POINTS_FOR_INFERENCE,
        "devices_monitored":    len(a.get_all_devices()),
        "model_path":           str(a.model_path),
        "resources_path":       str(a.resources_path),
    }


@app.post("/v4/add_message")
async def add_message(request: MessageRequest):
    a          = _get()
    has_enough = a.add_message(request.device_address, request.features.model_dump())
    stats      = a.get_device_stats(request.device_address)
    result     = None
    if has_enough:
        result = await a.analyze_device(request.device_address)
    return {
        "device_address":  request.device_address,
        "has_enough_data": has_enough,
        "buffer_size":     stats.get("messages_collected", 0),
        "analysis":        result,
    }


@app.get("/v4/analyze/{device_address}")
async def analyze_device(device_address: str):
    if not (device_address.startswith("0x") and len(device_address) == 42):
        raise HTTPException(status_code=400, detail="Adresse Ethereum invalide")
    return await _get().analyze_device(device_address.lower())


@app.get("/v4/stats/{device_address}")
async def get_device_stats(device_address: str):
    return _get().get_device_stats(device_address)

@app.get("/v4/devices")
async def get_all_devices():
    return {"devices": _get().get_all_devices()}

@app.get("/v4/anomalies")
async def get_recent_anomalies(limit: int = 20):
    return {"anomalies": _get().get_recent_anomalies(min(max(limit,1),200))}

@app.get("/v4/actions")
async def get_action_log(limit: int = 20):
    return {"actions": _get().get_action_log(min(max(limit,1),200))}

@app.get("/v4/stats")
async def get_global_stats():
    return _get().get_stats()

@app.get("/v4/health")
async def v4_health():
    return await health()

@app.get("/v4/debug/{device_address}")
async def debug_device(device_address: str):
    a = _get()
    s = a.get_device_stats(device_address)
    if "error" in s:
        raise HTTPException(status_code=404, detail=s["error"])
    return s


# ============================================================
# Point d'entrée
# ============================================================

if __name__ == "__main__":
    port = int(os.getenv("ML_PORT", "5001"))
    model_dir, res_dir = _find_model_path()

    print()
    print("=" * 70)
    print("🚀 IoT Network Analyzer V5 DEV v3.2")
    print(f"   Port           : {port}")
    print(f"   Backend URL    : {_BACKEND_URL}")
    print(f"   Modèle (.keras): {model_dir}")
    print(f"   Ressources     : {res_dir}")
    print(f"   TLS verify     : {os.getenv('TLS_VERIFY', 'true')}")
    print(f"   Debug ML       : {_DEBUG_ML}")
    print(f"   Threshold ×    : NO_SCALER_THRESH_FACTOR={_NO_SCALER_THRESH_FACTOR}")
    print(f"   Min points     : MIN_POINTS_FOR_INFERENCE={_MIN_POINTS_FOR_INFERENCE}")
    print()
    print("   ✅ FIX-PATH-5   : resources_dir remonte les ancêtres (scaler dans models/)")
    print("   ✅ FIX-SCALER-5 : copie auto scaler.json dans model_dir")
    print("   ✅ FIX-NORM-4   : z-score sur buffer complet")
    print("   ✅ FIX-THRESH-3 : facteur uniquement si scaler absent")
    print("=" * 70)
    print()

    uvicorn.run(app, host="0.0.0.0", port=port, log_level="info")
