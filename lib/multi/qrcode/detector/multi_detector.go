package detector

import (
	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/common"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/detector"
)

type MultiDetector struct {
	*detector.Detector
}

func NewMultiDetector(image *lib.BitMatrix) *MultiDetector {
	return &MultiDetector{
		detector.NewDetector(image),
	}
}

func (this *MultiDetector) DetectMulti(hints map[lib.DecodeHintType]interface{}) ([]*common.DetectorResult, error) {
	image := this.GetImage()
	resultPointCallback, _ := hints[lib.DecodeHintType_NEED_RESULT_POINT_CALLBACK].(lib.ResultPointCallback)

	finder := NewMultiFinderPatternFinder(image, resultPointCallback)
	infos, e := finder.FindMulti(hints)
	if e != nil || len(infos) == 0 {
		return nil, lib.WrapNotFoundException(e)
	}

	result := make([]*common.DetectorResult, 0)
	for _, info := range infos {
		r, e := this.ProcessFinderPatternInfo(info)
		if e != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}
