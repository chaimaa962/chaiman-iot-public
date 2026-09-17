import json, pickle, numpy as np
from pathlib import Path

pkl_path = Path("./export_3datasets_v5_final/scaler.pkl")
if not pkl_path.exists():
    pkl_path = Path("./scaler.pkl")

with open(pkl_path, 'rb') as f:
    scaler = pickle.load(f)

data = {
    "mean": scaler.mean_.tolist(),
    "scale": scaler.scale_.tolist(),
    "var": scaler.var_.tolist(),
    "n_samples_seen": int(getattr(scaler, 'n_samples_seen_', 0))
}

# Sauvegarder dans export_3datasets_v5_final/
out = Path("./export_3datasets_v5_final/scaler.json")
with open(out, 'w') as f:
    json.dump(data, f, indent=2)

print(f"✅ scaler.json généré : {out}")
print(f"   Features: {len(scaler.mean_)}")
print(f"   Mean: [{min(scaler.mean_):.4f}, {max(scaler.mean_):.4f}]")
print(f"   Scale: [{min(scaler.scale_):.4f}, {max(scaler.scale_):.4f}]")

# Copier aussi dans models/ (parent)
out2 = Path("./scaler.json")
with open(out2, 'w') as f:
    json.dump(data, f, indent=2)
print(f"✅ Copié aussi dans : {out2}")
