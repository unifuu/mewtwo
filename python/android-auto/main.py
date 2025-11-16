import time
import uiautomator2 as u2
import cv2
import numpy as np
import os
import matplotlib.pylab as plt
import pyautogui
import pytesseract
from PIL import Image

threshold = 0.8
app_name = "com.android.settings"
target_img = "gallery.png"
screen_size = (800, 600)
pcrd_running = False

# Connect to the device
d = u2.connect()

def launch_menu():
    # Press the home button to ensure on the home screen
    d.press("home")

    # Get the device screen size
    screen_width, screen_height = d.window_size()

    # Swipe from the bottom to the top to open the apps menu
    start_x = screen_width // 2
    start_y = int(screen_height * 0.85)
    end_x = screen_width // 2
    end_y = int(screen_height * 0.4)

    # Perform the swipe
    d.swipe(start_x, start_y, end_x, end_y, duration=0.1)

def launch_app(app_name):
    # Main loop to wait for the app to be in the foreground
    while True:
        if d.app_wait(app_name, front=True, timeout=1):
            if not pcrd_running:
                time.sleep(2)  # Wait before continuing
            pcrd_running = True
            break
        else:
            # Launch the app if it's not running
            d.session(app_name)
            pcrd_running = False
            continue

def detect_text():
    # Take a screenshot
    screenshot_path = 'screenshot.png'
    d.screenshot(screenshot_path)

    # Check if the file exists before trying to open it
    if not os.path.exists(screenshot_path):
        print(f"File does not exist: {screenshot_path}")
        return

    # Open the screenshot image
    try:
        image = Image.open(screenshot_path)
    except PermissionError as e:
        print(f"Permission error: {e}")
        return

    # Detect text with bounding boxes
    boxes = pytesseract.image_to_boxes(image)
    print("Detected text boxes: ", boxes)

def read_img(img):
    return cv2.imread(img, cv2.IMREAD_COLOR)

# Rotate the screen if in portrait mode
def rotate_screen_to_landscape():
    orientation = d.orientation

    if orientation == "natural":
        d.set_orientation("l")

def detect_image(img_path, threshold=threshold, new_size=screen_size):
    screen_shot = d.screenshot(format="opencv")

    # Convert the screenshot to a format that OpenCV can use (numpy array)
    screen_shot = np.array(screen_shot)
    screen_shot = cv2.cvtColor(screen_shot, cv2.COLOR_RGB2BGR)

    # Calculate new size based on the scale factor
    scale_factor = 0.4
    new_size = (int(screen_shot.shape[1] * scale_factor), int(screen_shot.shape[0] * scale_factor))

    # Resize the screenshot
    screen_shot = cv2.resize(screen_shot, new_size)

    # cv2.imshow('Detected Image', screen_shot)
    # cv2.waitKey(0)
    # cv2.destroyAllWindows()

    # Read the image to detect
    target_image = cv2.imread(img_path, cv2.IMREAD_COLOR)
    
    # Get the width and height of the target image
    h, w, _ = target_image.shape
    
    # Perform template matching
    result = cv2.matchTemplate(screen_shot, target_image, cv2.TM_CCOEFF_NORMED)
    
    # Find locations where the match quality exceeds the threshold
    locations = np.where(result >= threshold)

    if len(locations[0]) > 0:
        for point in zip(*locations[::-1]):  # Switch x and y coordinates
            # Draw a rectangle around the matched region
            cv2.rectangle(screen_shot, point, (point[0] + w, point[1] + h), (0, 255, 0), 2)
        
        # Show the result with the highlighted area
        cv2.imshow('Detected Image', screen_shot)
        cv2.waitKey(0)
        cv2.destroyAllWindows()
        return True
    
    else:
        print("Image not detected.")
        return False

def detect_image2(img_path, threshold=threshold):
    # Take a screenshot of the Android device
    screen_shot = d.screenshot(format="opencv")
    
    screen_shot = np.array(screen_shot)
    screen_shot = cv2.cvtColor(screen_shot, cv2.COLOR_RGB2BGR)

    # Read the target image
    target_image = read_img(img_path)
    target_image = cv2.cvtColor(target_image, cv2.COLOR_RGB2BGR)
    
    # Get the width and height of the target image
    h, w, _ = target_image.shape

    # Initialize a flag for detection
    detected = False

    # Define scales to test
    scales = np.linspace(0.1, 3, num=10)  # Scale from 10% to 300%

    for scale in scales:
        # Resize the target image
        resized_target = cv2.resize(target_image, (int(w * scale), int(h * scale)))

        # Perform template matching
        result = cv2.matchTemplate(screen_shot, resized_target, cv2.TM_CCOEFF_NORMED)

        # Find locations where the match quality exceeds the threshold
        locations = np.where(result >= threshold)

        if len(locations[0]) > 0:
            detected = True
            for point in zip(*locations[::-1]):  # Switch x and y coordinates
                # Draw a rectangle around the matched region
                cv2.rectangle(screen_shot, point, (point[0] + int(w * scale), point[1] + int(h * scale)), (0, 255, 0), 2)

                # Calculate the center of the matched region and convert to int
                center_x = int(point[0] + w // 2)
                center_y = int(point[1] + h // 2)
            
                # Simulate a tap at the center of the matched region
                d.click(center_x, center_y)

                # Show the result with the highlighted area
                # cv2.imshow('Detected Image', screen_shot)
                # cv2.waitKey(0)
                # cv2.destroyAllWindows()

                # break  # Break if a match is found
                return True

    if not detected:
        print("Image not detected.")
        return False

    return True

def detect_image_on_screen(image_path, threshold=threshold):
    # Take a screenshot of the Android simulator
    screenshot = pyautogui.screenshot()
    
    # Convert the screenshot to a format that OpenCV can use (numpy array)
    screenshot = np.array(screenshot)
    screenshot = cv2.cvtColor(screenshot, cv2.COLOR_RGB2BGR)

    # Read the image you want to detect
    target_image = cv2.imread(image_path, cv2.IMREAD_COLOR)
    
    # Get the width and height of the target image
    h, w, _ = target_image.shape
    
    # Perform template matching
    result = cv2.matchTemplate(screenshot, target_image, cv2.TM_CCOEFF_NORMED)
    
    # Find locations where the match quality exceeds the threshold
    locations = np.where(result >= threshold)
    
    if len(locations[0]) > 0:
        for point in zip(*locations[::-1]):  # Switch x and y coordinates
            # Draw a rectangle around the matched region
            cv2.rectangle(screenshot, point, (point[0] + w, point[1] + h), (0, 255, 0), 2)
        
        # Show the result with the highlighted area
        cv2.imshow('Detected Image', screenshot)
        cv2.waitKey(0)
        cv2.destroyAllWindows()
        return True
    else:
        print("Image not detected.")
        return False

launch_app(app_name)

# detect_image2(target_img)

# detect_text()

# launch_menu()