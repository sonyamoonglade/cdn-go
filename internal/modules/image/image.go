package image

import (
	"bytes"

	"animakuro/cdn/internal/modules/errors"
	"animakuro/cdn/internal/modules/types"
	"github.com/h2non/bimg"
	"github.com/pkg/errors"
)

const (
	ModuleName         = "image"
	Webp               = "webp"
	Resized            = "resized"
	TrueStr            = "true"
	FalseStr           = "false"
	YAxisResizePercent = 0.2114
)

var (
	ErrInvalidArgs = errors.New("invalid arguments")
)

func Init() *types.Module {
	m := &types.Module{
		Name:                  ModuleName,
		Resolvers:             make(map[string]types.ResolverFunc),
		Defaults:              make(types.Defaults),
		AllowedResolverValues: make(map[string][]string),
	}

	//Set resolvers
	m.Resolvers[Webp] = webpfn
	m.Resolvers[Resized] = resizedfn

	//Set defaults
	m.Defaults[Webp] = FalseStr
	m.Defaults[Resized] = FalseStr

	m.AllowedResolverValues[Webp] = []string{TrueStr, FalseStr}
	m.AllowedResolverValues[Resized] = []string{TrueStr, FalseStr}

	return m
}

func webpfn(buff *bytes.Buffer, arg interface{}) error {

	if arg != TrueStr {
		return nil
	}

	img := bimg.NewImage(buff.Bytes())

	newimg, err := img.Convert(bimg.WEBP)
	if err != nil {
		return module_errors.WrapInternal(err, "image.webpfn.img.Convert")
	}

	(*buff).Reset()
	(*buff).Write(newimg)

	return nil
}

func resizedfn(buff *bytes.Buffer, arg interface{}) error {
	if arg != TrueStr {
		return nil
	}

	img := bimg.NewImage(buff.Bytes())

	size, err := img.Size()
	if err != nil {
		return module_errors.WrapInternal(err, "image.resizedfn.img.Size")
	}

	H := float64(size.Height)
	hmin := H * YAxisResizePercent
	hmax := H * (1 - 2*YAxisResizePercent)

	newimg, err := img.Extract(int(hmin), 0, size.Width, int(hmax))
	if err != nil {
		return module_errors.WrapInternal(err, "image.resizedfn.img.Extract")
	}

	(*buff).Reset()
	(*buff).Write(newimg)

	return nil
}
