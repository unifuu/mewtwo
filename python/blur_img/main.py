from PIL import Image, ImageFilter

img = Image.open("input.jpg")
blurred = img.filter(ImageFilter.GaussianBlur(3))
blurred.save("blurred.jpg")
