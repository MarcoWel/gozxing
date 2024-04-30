package multi

import (
	"github.com/MarcoWel/gozxing"
	"github.com/MarcoWel/gozxing/aztec"
	"github.com/MarcoWel/gozxing/datamatrix"
	"github.com/MarcoWel/gozxing/oned"
	"github.com/MarcoWel/gozxing/qrcode"
)

type MultiFormatReader struct {
	hints   map[gozxing.DecodeHintType]interface{}
	readers []gozxing.Reader
}

func NewMultiFormatReader() *MultiFormatReader {
	return &MultiFormatReader{}
}

func (r *MultiFormatReader) DecodeWithoutHints(image *gozxing.BinaryBitmap) (*gozxing.Result, error) {
	r.SetHints(nil)
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) Decode(image *gozxing.BinaryBitmap, hints map[gozxing.DecodeHintType]interface{}) (*gozxing.Result, error) {
	r.SetHints(hints)
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) DecodeWithState(image *gozxing.BinaryBitmap) (*gozxing.Result, error) {
	if r.readers == nil {
		r.SetHints(nil)
	}
	return r.DecodeInternal(image)
}

func (r *MultiFormatReader) SetHints(hints map[gozxing.DecodeHintType]interface{}) {
	r.hints = hints

	var formats []gozxing.BarcodeFormat
	if hintFormats, ok := hints[gozxing.DecodeHintType_POSSIBLE_FORMATS]; ok {
		formats = hintFormats.([]gozxing.BarcodeFormat)
	}

	var readers []gozxing.Reader
	for _, format := range formats {
		switch format {
		case gozxing.BarcodeFormat_QR_CODE:
			readers = append(readers, qrcode.NewQRCodeReader())
		case gozxing.BarcodeFormat_DATA_MATRIX:
			readers = append(readers, datamatrix.NewDataMatrixReader())
		case gozxing.BarcodeFormat_AZTEC:
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

func (r *MultiFormatReader) DecodeInternal(image *gozxing.BinaryBitmap) (*gozxing.Result, error) {
	for _, reader := range r.readers {
		result, err := reader.Decode(image, r.hints)
		if err == nil {
			return result, nil
		}
	}
	return nil, gozxing.NewNotFoundException()
}
