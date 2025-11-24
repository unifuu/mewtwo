from PIL import Image

# Open JPG
img = Image.open("input.png")

# Optional: resize to standard favicon sizes
sizes = [(16,16), (32,32), (48,48), (64,64)]
img.save("favicon.ico", sizes=sizes)
