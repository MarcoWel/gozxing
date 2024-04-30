package gozxing

import (
	"github.com/MarcoWel/gozxing/aztec"
	"github.com/MarcoWel/gozxing/datamatrix"
	"github.com/MarcoWel/gozxing/oned"
	"github.com/MarcoWel/gozxing/qrcode"
)

type MultiFormatReader struct {
	hints   map[DecodeHintType]interface{}
	readers []Reader
}

func NewMultiFormatReader() *MultiFormatReader {
	return &MultiFormatReader{}
}

func (r *MultiFormatReader) Decode(image *BinaryBitmap) (*Result, error) {
	r.SetHints(nil)
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) DecodeWithHints(image *BinaryBitmap, hints map[DecodeHintType]interface{}) (*Result, error) {
	r.SetHints(hints)
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) DecodeWithState(image *BinaryBitmap) (*Result, error) {
	if r.readers == nil {
		r.SetHints(nil)
	}
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) SetHints(hints map[DecodeHintType]interface{}) {
	r.hints = hints

	var formats []BarcodeFormat
	if hintFormats, ok := hints[DecodeHintType_POSSIBLE_FORMATS]; ok {
		formats = hintFormats.([]BarcodeFormat)
	}

	var readers []Reader
	for _, format := range formats {
		switch format {
		case BarcodeFormat_QR_CODE:
			readers = append(readers, qrcode.NewQRCodeReader())
		case BarcodeFormat_DATA_MATRIX:
			readers = append(readers, datamatrix.NewDataMatrixReader())
		case BarcodeFormat_AZTEC:
			readers = append(readers, aztec.NewAztecReader())
		// case BarcodeFormat_PDF_417:
		// 	readers = append(readers, pdf417.NewPDF417Reader())
		// case BarcodeFormat_MAXICODE:
		// 	readers = append(readers, maxicode.NewMaxicodeReader())
		default:
			readers = append(readers, oned.NewMultiFormatOneDReader(hints))
		}
	}

	if len(readers) == 0 {
		// Default to all formats if none specified
		readers = append(readers, oned.NewMultiFormatOneDReader(hints))
		readers = append(readers, qrcode.NewQRCodeReader())
		readers = append(readers, datamatrix.NewDataMatrixReader())
		readers = append(readers, aztec.NewAztecReader())
		// readers = append(readers, pdf417.NewReader())
		// readers = append(readers, maxicode.NewReader())
	}

	r.readers = readers
}

func (r *MultiFormatReader) Reset() {
	for _, reader := range r.readers {
		reader.Reset()
	}
}

func (r *MultiFormatReader) DecodeInternal(image *BinaryBitmap) (*Result, error) {
	for _, reader := range r.readers {
		result, err := reader.Decode(image, r.hints)
		if err == nil {
			return result, nil
		}
	}
	return nil, NewNotFoundException()
}
