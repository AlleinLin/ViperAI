package service

import (
	"io"
	"log"
	"mime/multipart"

	imageutil "viperai/internal/pkg/image"
)

type ImageService struct {
	modelPath string
	labelPath string
	inputH    int
	inputW    int
}

func NewImageService(modelPath, labelPath string, inputH, inputW int) *ImageService {
	return &ImageService{
		modelPath: modelPath,
		labelPath: labelPath,
		inputH:    inputH,
		inputW:    inputW,
	}
}

func (s *ImageService) Recognize(file *multipart.FileHeader) (string, error) {
	recognizer, err := imageutil.NewRecognizer(s.modelPath, s.labelPath, s.inputH, s.inputW)
	if err != nil {
		log.Println("Failed to create image recognizer:", err)
		return "", err
	}
	defer recognizer.Close()

	src, err := file.Open()
	if err != nil {
		log.Println("Failed to open file:", err)
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		log.Println("Failed to read file:", err)
		return "", err
	}

	return recognizer.PredictFromBuffer(buf)
}
