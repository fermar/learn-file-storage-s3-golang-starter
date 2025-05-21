package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

func getVideoAspectRatio(filepath string) (string, error) {
	type ffprobeOutput struct {
		Streams []struct {
			Index              int    `json:"index"`
			CodecName          string `json:"codec_name,omitempty"`
			CodecLongName      string `json:"codec_long_name,omitempty"`
			Profile            string `json:"profile,omitempty"`
			CodecType          string `json:"codec_type"`
			CodecTagString     string `json:"codec_tag_string"`
			CodecTag           string `json:"codec_tag"`
			Width              int    `json:"width,omitempty"`
			Height             int    `json:"height,omitempty"`
			CodedWidth         int    `json:"coded_width,omitempty"`
			CodedHeight        int    `json:"coded_height,omitempty"`
			ClosedCaptions     int    `json:"closed_captions,omitempty"`
			FilmGrain          int    `json:"film_grain,omitempty"`
			HasBFrames         int    `json:"has_b_frames,omitempty"`
			SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
			DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
			PixFmt             string `json:"pix_fmt,omitempty"`
			Level              int    `json:"level,omitempty"`
			ColorRange         string `json:"color_range,omitempty"`
			ColorSpace         string `json:"color_space,omitempty"`
			ColorTransfer      string `json:"color_transfer,omitempty"`
			ColorPrimaries     string `json:"color_primaries,omitempty"`
			ChromaLocation     string `json:"chroma_location,omitempty"`
			FieldOrder         string `json:"field_order,omitempty"`
			Refs               int    `json:"refs,omitempty"`
			IsAvc              string `json:"is_avc,omitempty"`
			NalLengthSize      string `json:"nal_length_size,omitempty"`
			ID                 string `json:"id"`
			RFrameRate         string `json:"r_frame_rate"`
			AvgFrameRate       string `json:"avg_frame_rate"`
			TimeBase           string `json:"time_base"`
			StartPts           int    `json:"start_pts"`
			StartTime          string `json:"start_time"`
			DurationTs         int    `json:"duration_ts"`
			Duration           string `json:"duration"`
			BitRate            string `json:"bit_rate,omitempty"`
			BitsPerRawSample   string `json:"bits_per_raw_sample,omitempty"`
			NbFrames           string `json:"nb_frames"`
			ExtradataSize      int    `json:"extradata_size"`
			Disposition        struct {
				Default         int `json:"default"`
				Dub             int `json:"dub"`
				Original        int `json:"original"`
				Comment         int `json:"comment"`
				Lyrics          int `json:"lyrics"`
				Karaoke         int `json:"karaoke"`
				Forced          int `json:"forced"`
				HearingImpaired int `json:"hearing_impaired"`
				VisualImpaired  int `json:"visual_impaired"`
				CleanEffects    int `json:"clean_effects"`
				AttachedPic     int `json:"attached_pic"`
				TimedThumbnails int `json:"timed_thumbnails"`
				NonDiegetic     int `json:"non_diegetic"`
				Captions        int `json:"captions"`
				Descriptions    int `json:"descriptions"`
				Metadata        int `json:"metadata"`
				Dependent       int `json:"dependent"`
				StillImage      int `json:"still_image"`
			} `json:"disposition"`
			Tags struct {
				Language    string `json:"language"`
				HandlerName string `json:"handler_name"`
				VendorID    string `json:"vendor_id"`
				Encoder     string `json:"encoder"`
				Timecode    string `json:"timecode"`
			} `json:"tags,omitempty"`
			SampleFmt      string `json:"sample_fmt,omitempty"`
			SampleRate     string `json:"sample_rate,omitempty"`
			Channels       int    `json:"channels,omitempty"`
			ChannelLayout  string `json:"channel_layout,omitempty"`
			BitsPerSample  int    `json:"bits_per_sample,omitempty"`
			InitialPadding int    `json:"initial_padding,omitempty"`
			Tags0          struct {
				Language    string `json:"language"`
				HandlerName string `json:"handler_name"`
				VendorID    string `json:"vendor_id"`
			} `json:"tags0,omitempty"`
			Tags1 struct {
				Language    string `json:"language"`
				HandlerName string `json:"handler_name"`
				Timecode    string `json:"timecode"`
			} `json:"tags1,omitempty"`
		} `json:"streams"`
	}

	slog.Info("get aspect", "file", filepath)
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filepath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	ffpOut := ffprobeOutput{}
	err = json.Unmarshal(out.Bytes(), &ffpOut)
	if err != nil {
		return "", err
	}

	// slog.Info("salida ffprobe:", "Width", ffpOut.Streams[0].Width)
	// slog.Info("salida ffprobe:", "Heigth", ffpOut.Streams[0].Height)
	ratio := float32(ffpOut.Streams[0].Width) / float32(ffpOut.Streams[0].Height)
	//
	// slog.Info("salida ffprobe:", "ratio", ratio)

	if ratio > 1.7 && ratio < 1.8 {
		return "16:9", nil
	}

	if ratio > 0.5 && ratio < 0.6 {
		return "9:16", nil
	}
	return "other", nil
}

func processVideoForFastStart(filepath string) (string, error) {
	procFilepath := fmt.Sprintf("%s.processing", filepath)
	cmd := exec.Command(
		"ffmpeg",
		"-i",
		filepath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		procFilepath,
	)

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return procFilepath, nil
}

// func generatePresignedURL(
// 	s3Client *s3.Client,
// 	bucket, key string,
// 	expireTime time.Duration,
// ) (string, error) {
// 	preSignCli := s3.NewPresignClient(s3Client)
// 	pscliParams := s3.GetObjectInput{Bucket: &bucket, Key: &key}
// 	pshttpr, err := preSignCli.PresignGetObject(
// 		context.Background(),
// 		&pscliParams,
// 		s3.WithPresignExpires(expireTime),
// 	)
// 	if err != nil {
// 		return "", err
// 	}
// 	return pshttpr.URL, nil
// }
//
// func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
// 	slog.Warn(
// 		"entreada a dbvideo to signed video",
// 		"video URL",
// 		video.VideoURL,
// 	)
// 	if video.VideoURL == nil {
// 		return video, nil
// 	}
//
// 	bucketKey := strings.Split(*video.VideoURL, ",")
//
// 	slog.Warn(
// 		"entreada a dbvideo to signed video",
// 		"bucketkey",
// 		bucketKey,
// 	)
//
// 	if len(bucketKey) == 2 {
// 		psVideoURL, err := generatePresignedURL(
// 			cfg.s3Client,
// 			bucketKey[0],
// 			bucketKey[1],
// 			2*time.Minute,
// 		)
// 		if err != nil {
// 			return video, err
// 		}
//
// 		retVideo := video
// 		retVideo.VideoURL = &psVideoURL
// 		return retVideo, nil
// 	}
// 	return database.Video{}, nil
// }
