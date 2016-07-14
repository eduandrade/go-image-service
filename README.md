# go-image-service

Rest API written in GO for uploading images and dynamically dimensioning them according to the rest parameters.

Format:
http://localhost:8080/img/<image id>/<width>/<height>

Example:
http://localhost:8080/img/testid/300/150
Return the image with id testid with 300px width and 150px height.
