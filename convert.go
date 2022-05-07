package main

import (
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func imToWebp(f string) (string, error) {
	pathOut := f + ".webp"
	args := []string{}
	var bin string
	if runtime.GOOS == "linux" {
		bin = "convert"
	} else {
		bin = "magick"
		args = append(args, "convert")
	}
	args = append(args, "-resize", "512x512", f, pathOut)

	cmd := exec.Command(bin, args...)
	err := cmd.Run()

	return pathOut, err
}

func imToPng(f string) (string, error) {
	pathOut := f + ".png"
	args := []string{}
	var bin string
	if runtime.GOOS == "linux" {
		bin = "convert"
	} else {
		bin = "magick"
		args = append(args, "convert")
	}
	args = append(args, f, pathOut)

	cmd := exec.Command(bin, args...)
	err := cmd.Run()

	return pathOut, err
}

func ffToWebm(f string) (string, error) {
	pathOut := f + ".webm"
	bin := "ffmpeg"
	args := []string{"-hide_banner", "-i", f,
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
		"-c:v", "libvpx-vp9", "-cpu-used", "8", "-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
		"-to", "00:00:02.900", "-an", "-y", pathOut}

	output, err := exec.Command(bin, args...).CombinedOutput()
	log.Traceln(string(output))
	// err := cmd.Run()
	return pathOut, err
}

func ffToWebmSafe(f string) (string, error) {
	pathOut := f + ".webm"
	bin := "ffmpeg"
	args := []string{"-hide_banner", "-i", f,
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
		"-c:v", "libvpx-vp9", "-cpu-used", "8", "-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
		"-to", "00:00:02.500", "-r", "30", "-an", "-y", pathOut}

	cmd := exec.Command(bin, args...)
	err := cmd.Run()
	return pathOut, err
}

func ffToGif(f string) (string, error) {
	var decoder []string
	var args []string
	if strings.HasSuffix(f, ".webm") {
		decoder = []string{"-c:v", "libvpx-vp9"}
	}
	pathOut := f + ".gif"
	bin := "ffmpeg"
	args = append(args, decoder...)
	args = append(args, "-i", f, "-hide_banner",
		"-lavfi", "scale=256:256:force_original_aspect_ratio=decrease,split[a][b];[a]palettegen=reserve_transparent=on:transparency_color=ffffff[p];[b][p]paletteuse",
		"-loglevel", "error", pathOut)

	cmd := exec.Command(bin, args...)
	err := cmd.Run()
	return pathOut, err
}

func imStackToWebp(base string, overlay string) (string, error) {
	args := []string{}
	fOut := base + ".composite.webp"
	var bin string
	if runtime.GOOS == "linux" {
		bin = "convert"
	} else {
		bin = "magick"
		args = append(args, "convert")
	}
	args = append(args, base, overlay, "-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-composite",
		"-define", "webp:lossless=true", fOut)
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Error("IM stack ERROR!")
		log.Errorln(out)
		return "", err
	} else {
		return fOut, nil
	}
}
