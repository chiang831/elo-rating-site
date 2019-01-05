# python 3
# Before running the code, follow the instruction given in
#https://googleapis.github.io/google-cloud-python/latest/vision/
# pip install virtualenv
# virtualenv <your-env>
# source <your-env>/bin/activate
# <your-env>/bin/pip install google-cloud-vision
# <your-env>/bin/pip install statistics
# 
# also, follow https://cloud.google.com/docs/authentication/getting-started and set the environment variable.
# 
# Examples:
# detect_text_uri("gs://elo-site-201802-test/user_upload/game1.PNG")
# result = TTAResult("game1.txt")
# result = TTAResult("gs://elo-site-201802-test/user_upload/IMG_0405.PNG")
# result=TTAResult("IMG_0405.txt")
# result = TTAResult("gs://elo-site-201802-test/user_upload/Screenshot_20181231-121516_Through the Ages.jpg")
# result = TTAResult("Screenshot_20181231-121516_Through the Ages.txt")
# result = TTAResult("gs://elo-site-201802-test/user_upload/cookie.PNG")
# result = TTAResult("cookie.txt")
# result.save_as_json("cookie.txt")
# print(result.player_ranking())
# print(result.player_score())
# print(result.player_order())
import io
import os
import json
import numpy as np
import collections
from google.protobuf.json_format import MessageToJson
from google.protobuf.json_format import Parse
from google.cloud.vision_v1.types import AnnotateImageResponse
from statistics import mean
from difflib import get_close_matches


# Imports the Google Cloud client library
from google.cloud import vision
from google.cloud.vision import types

def vertices_across_column_(vertices, position):
    has_left_vertices = False
    has_right_vertices = False
    for vertex in vertices:
        if vertex.x <= position:
            has_left_vertices = True
        if vertex.x >= position:
            has_right_vertices = True 
        if has_left_vertices and has_right_vertices:
            return True
    return False


    

TextWithAvgY_ = collections.namedtuple("TextWithAvgY_", ["avg_y", "text"])

RANKING_REL_POSITIONS_ = np.linspace(0.45, 0.55, 11)
SCORE_REL_POSITIONS_ = np.linspace(0.55, 0.65, 11)
ORDER_REL_POSITIONS_ = np.linspace(0.87, 0.97, 11)
PLAYERS_ = ["6bro", "Neil.W", "chiang831", "pg30123", "stan619", "cookieben", "DHYellow"]

class TTAResult(object):
    
    def annotated_from_uri_(self, uri):
        """Detects text in the file located in Google Cloud Storage or on the Web.
        """
        client = vision.ImageAnnotatorClient()
        image = vision.types.Image()
        image.source.image_uri = uri
    
        self.response_= client.text_detection(image=image)


    def save_as_json(self, path):
        serialized = MessageToJson(self.response_)
        with open(path, 'w') as outfile:  
            json.dump(serialized, outfile)
    
    def read_json_file_(self, path):
        with open(path,'r') as json_data:
            data = json.load(json_data)
        self.response_ = Parse(data, AnnotateImageResponse())

    def __init__(self, path):
        if path.startswith("gs://"):
            self.annotated_from_uri_(path)
        else:
            self.read_json_file_(path)
        self.player_ranking_ = None
        self.player_score_ = None
        self.player_order_ = None

    
    def sorted_text_across_column_(self, rel_position):
        """
        Returns a list of text across a given column, sorted from top to bottom.
        Args:
            rel_position: the relative rel_position of the column. this should be between 0 and 1
        """
        if rel_position < 0 or rel_position > 1:
            return []
        width = self.response_.full_text_annotation.pages[0].width
        column_value = rel_position * width
        texts = self.response_.text_annotations
        # A list of TextWithAvgY_ for texts across 
        text_in_range = []
        for text in texts:
            vertices = text.bounding_poly.vertices
            if vertices_across_column_(vertices, column_value):
                text_in_range.append(TextWithAvgY_(avg_y=mean([vertex.y for vertex in vertices]), text=text.description))
                
        text_in_range.sort(key=lambda x:x.avg_y)
        return [text_with_avg_y.text for text_with_avg_y in text_in_range ]
                    
    def player_ranking(self):
        if self.player_ranking_ is not None:
            return self.player_ranking_
        self.player_ranking_ = []
        for rel_position in RANKING_REL_POSITIONS_:
            result = self.sorted_text_across_column_(rel_position)
            close_matches = [get_close_matches(name, possibilities=PLAYERS_)[0] for name in result if len(get_close_matches(name, possibilities=PLAYERS_))>0]
            if len(close_matches)==0:
                continue
            # Often the name of the game is of the form 6bro's game, and hence the first close match of a player id may be a false positive.
            if close_matches.count(close_matches[0])>1:
                close_matches.remove(close_matches[0])
            if len(self.player_ranking_) < len(close_matches):
                self.player_ranking_ =close_matches
        return self.player_ranking_
    
    def player_score(self):
        if self.player_score_ is not None:
            return self.player_score_
        self.player_score_ =[]
        for rel_position in SCORE_REL_POSITIONS_:
            result = self.sorted_text_across_column_(rel_position)
            candidate = [score for score in result if score.isdigit()]
            if len(self.player_score_)<len(candidate):
                self.player_score_ = candidate
        return self.player_score_
        
    def player_order(self):
        if self.player_order_ is not None:
            return self.player_order_
        self.player_order_ = []
        for rel_position in ORDER_REL_POSITIONS_:
            result = self.sorted_text_across_column_(rel_position)
            close_matches = [get_close_matches(name, possibilities=PLAYERS_)[0] for name in result if len(get_close_matches(name, possibilities=PLAYERS_))>0]
            if len(self.player_order_) < len(close_matches):
                self.player_order_ =close_matches
        return self.player_order_
