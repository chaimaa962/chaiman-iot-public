import sys
import os
import tensorflow as tf
import tf2onnx
import onnx

def main():
    if len(sys.argv) != 3:
        print("Usage: convert.py <input_model.keras> <output_model.onnx>")
        sys.exit(1)
    
    input_path = sys.argv[1]
    output_path = sys.argv[2]
    
    if not os.path.exists(input_path):
        print(f"❌ Fichier non trouvé: {input_path}")
        sys.exit(1)
    
    print(f"📥 Chargement: {input_path}")
    model = tf.keras.models.load_model(input_path)
    
    print(f"🔄 Conversion vers ONNX...")
    spec = (tf.TensorSpec(model.input_shape, tf.float32, name="input"),)
    model_proto, _ = tf2onnx.convert.from_keras(
        model, 
        input_signature=spec, 
        opset=13
    )
    
    print(f"💾 Sauvegarde: {output_path}")
    onnx.save(model_proto, output_path)
    print(f"✅ Conversion terminée!")

if __name__ == "__main__":
    main()
