package multi

import (
	"fmt"

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
	fmt.Println("Decoding image at", xOffset, yOffset, "depth", currentDepth)
	if currentDepth > maxDepth {
		fmt.Println("Too deep, returning nothing")
		return nil
	}

	result, err := r.delegate.Decode(image, hints)
	if err != nil {
		if _, ok := err.(gozxing.ReaderException); ok {
			fmt.Println("ReaderException, returning nothing")
			return nil // ignore ReaderException
		}
		return err
	}

	alreadyFound := false
	for _, existingResult := range *results {
		if existingResult.GetText() == result.GetText() {
			fmt.Println("Already found", result.GetText())
			alreadyFound = true
			break
		}
	}

	if !alreadyFound {
		fmt.Println("Appending result", result.GetText())
		translatedResult := translateResultPoints(result, xOffset, yOffset)
		*results = append(*results, translatedResult)
	}

	resultPoints := result.GetResultPoints()
	if len(resultPoints) == 0 {
		return nil
	}

	width := image.GetWidth()
	height := image.GetHeight()
	var minX, minY = width, height
	var maxX, maxY = 0, 0

	for _, point := range resultPoints {
		x, y := int(point.GetX()), int(point.GetY())
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

	// Decode left of barcode
	if minX > minDimensionToRecur {
		subImage, _ := image.Crop(0, 0, minX, height)
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset, currentDepth+1)
	}
	// Decode above barcode
	if minY > minDimensionToRecur {
		subImage, _ := image.Crop(0, 0, width, minY)
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset, currentDepth+1)
	}
	// Decode right of barcode
	if maxX < width-minDimensionToRecur {
		subImage, _ := image.Crop(maxX, 0, width-maxX, height)
		r.doDecodeMultiple(subImage, hints, results, xOffset+maxX, yOffset, currentDepth+1)
	}
	// Decode below barcode
	if maxY < height-minDimensionToRecur {
		subImage, _ := image.Crop(0, maxY, width, height-maxY)
		r.doDecodeMultiple(subImage, hints, results, xOffset, yOffset+maxY, currentDepth+1)
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
