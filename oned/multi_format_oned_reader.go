package oned

import (
	"github.com/MarcoWel/gozxing"
)

type MultiFormatOneDReader struct {
	*OneDReader
	readers []gozxing.Reader
}

func NewMultiFormatOneDReader(hints map[gozxing.DecodeHintType]interface{}) gozxing.Reader {
	possibleFormats, ok := hints[gozxing.DecodeHintType_POSSIBLE_FORMATS].([]gozxing.BarcodeFormat)
	var readers []gozxing.Reader

	if ok && len(possibleFormats) > 0 {
		for _, format := range possibleFormats {
			switch format {
			case gozxing.BarcodeFormat_EAN_13, gozxing.BarcodeFormat_UPC_A, gozxing.BarcodeFormat_EAN_8, gozxing.BarcodeFormat_UPC_E:
				readers = append(readers, NewMultiFormatUPCEANReader(hints))
			case gozxing.BarcodeFormat_CODE_39:
				//useCode39CheckDigit := hints[gozxing.DecodeHintType_POSSIBLE_FORMATS] != nil
				readers = append(readers, NewCode39Reader())
			case gozxing.BarcodeFormat_CODE_93:
				readers = append(readers, NewCode93Reader())
			case gozxing.BarcodeFormat_CODE_128:
				readers = append(readers, NewCode128Reader())
			case gozxing.BarcodeFormat_ITF:
				readers = append(readers, NewITFReader())
				// case gozxing.BarcodeFormat_CODABAR:
				// 	readers = append(readers, NewCondabarReader())
				// case gozxing.BarcodeFormat_RSS_14:
				// 	readers = append(readers, rss.NewRSS14Reader())
				// case gozxing.BarcodeFormat_RSS_EXPANDED:
				// 	readers = append(readers, rss.NewRSSExpandedReader())
			}
		}
	}

	if len(readers) == 0 {
		readers = append(readers, NewMultiFormatUPCEANReader(hints))
		readers = append(readers, NewCode39Reader())
		readers = append(readers, NewCode93Reader())
		readers = append(readers, NewCode128Reader())
		readers = append(readers, NewITFReader())
		// readers = append(readers, NewCondabarReader())
		// readers = append(readers, rss.NewRSS14Reader())
		// readers = append(readers, rss.NewRSSExpandedReader())
	}

	this := &MultiFormatOneDReader{
		readers: readers,
	}
	this.OneDReader = NewOneDReader(this)
	return this
}

func (r *MultiFormatOneDReader) DecodeRow(rowNumber int, row *gozxing.BitArray, hints map[gozxing.DecodeHintType]interface{}) (*gozxing.Result, error) {
	for _, reader := range r.readers {
		decoder := reader.(RowDecoder)
		result, err := decoder.DecodeRow(rowNumber, row, hints)
		if err == nil {
			return result, nil
		}
	}
	return nil, gozxing.NewNotFoundException()
}

func (r *MultiFormatOneDReader) Reset() {
	for _, reader := range r.readers {
		reader.Reset()
	}
}
