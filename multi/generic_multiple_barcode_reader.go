package multi

import (
	"github.com/MarcoWel/gozxing"
)

const (
	minDimensionToRecur = 100
	maxDepth            = 4
)

type GenericMultipleBarcodeReader struct {
	delegate gozxing.Reader
}

func NewGenericMultipleBarcodeReader(delegate gozxing.Reader) *GenericMultipleBarcodeReader {
	return &GenericMultipleBarcodeReader{delegate: delegate}
}

func (r *GenericMultipleBarcodeReader) DecodeMultiple(image *gozxing.BinaryBitmap, hints map[gozxing.DecodeHintType]interface{}) ([]*gozxing.Result, error) {
	if hints == nil {
		hints = make(map[gozxing.DecodeHintType]interface{})
	}
	results := make([]*gozxing.Result, 0)
	err := r.doDecodeMultiple(image, hints, &results, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, gozxing.NewNotFoundException()
	}
	return results, nil
}

func (r *GenericMultipleBarcodeReader) doDecodeMultiple(image *gozxing.BinaryBitmap, hints map[gozxing.DecodeHintType]interface{}, results *[]*gozxing.Result, xOffset, yOffset, currentDepth int) error {
	if currentDepth > maxDepth {
		return nil
	}

	result, err := r.delegate.Decode(image, hints)
	if err != nil {
		if _, ok := err.(gozxing.ReaderException); ok {
			return nil // ignore ReaderException
		}
		return err
	}

	alreadyFound := false
	for _, existingResult := range *results {
		if existingResult.GetText() == result.GetText() {
			alreadyFound = true
			break
		}
	}

	if !alreadyFound {
		translatedResult := translateResultPoints(result, xOffset, yOffset)
		*results = append(*results, translatedResult)
	}

	resultPoints := result.GetResultPoints()
	if resultPoints == nil || len(resultPoints) == 0 {
		return nil
	}

	width := image.GetWidth()
	height := image.GetHeight()
	var minX, minY = float64(width), float64(height)
	var maxX, maxY float64 = 0, 0

	for _, point := range resultPoints {
		x, y := point.GetX(), point.GetY()
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	if minX > minDimensionToRecur {
		subImage, _ := image.Crop(0, 0, int(minX), height)
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset, currentDepth+1)
	}
	if minY > minDimensionToRecur {
		subImage, _ := image.Crop(0, 0, width, int(minY))
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset, currentDepth+1)
	}
	if maxX < float64(width)-minDimensionToRecur {
		subImage, _ := image.Crop(int(maxX), 0, width-int(maxX), height)
		r.doDecodeMultiple(subImage, hints, results, xOffset+int(maxX), yOffset, currentDepth+1)
	}
	if maxY < float64(height)-minDimensionToRecur {
		subImage, _ := image.Crop(0, int(maxY), width, height-int(maxY))
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset+int(maxY), currentDepth+1)
	}

	return nil
}

func translateResultPoints(result *gozxing.Result, xOffset, yOffset int) *gozxing.Result {
	oldResultPoints := result.GetResultPoints()
	newResultPoints := make([]gozxing.ResultPoint, len(oldResultPoints))
	for i, oldPoint := range oldResultPoints {
		newResultPoints[i] = gozxing.NewResultPoint(oldPoint.GetX()+float64(xOffset), oldPoint.GetY()+float64(yOffset))
	}
	return gozxing.NewResult(result.GetText(), result.GetRawBytes(), newResultPoints, result.GetBarcodeFormat())
}
