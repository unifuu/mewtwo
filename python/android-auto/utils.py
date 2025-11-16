import cv2
import numpy as np

threshold = 0.8

class UIDetecter:
    @staticmethod
    def detect_imgs(screen_shot, imgs):
        centers = []

        # Range the images
        for img in imgs:
            # Read the image to detect
            t = read_img(img)

            # Get the width and height of the target image
            h, w = t.shape

            # Perform template matching
            result = cv2.matchTemplate(screen_shot, t, cv2.TM_CCOEFF_NORMED)

            # Find locations where match quality exceeds the threshold
            locations = np.where(result >= threshold)

            if len(locations[0]) > 0:
                for point in zip(*locations[::-1]):
                    center_x = int(point[0] + w // 2)
                    center_y = int(point[1] + h // 2)

                    centers.append([center_x, center_y])
        return centers