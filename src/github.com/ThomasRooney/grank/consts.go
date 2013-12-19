package main

const (
	QUALITY_HIGHRES = "highres"
	QUALITY_HD1080  = "hd1080"
	QUALITY_HD720   = "hd720"
	QUALITY_LARGE   = "large"
	QUALITY_MEDIUM  = "medium"
	QUALITY_SMALL   = "small"
	QUALITY_MIN     = "min"
	QUALITY_MAX     = "max"
	QUALITY_UNKNOWN = "unknown"

	FORMAT_MP4     = "mp4"
	FORMAT_WEBM    = "webm"
	FORMAT_FLV     = "flv"
	FORMAT_3GP     = "3ggp"
	FORMAT_UNKNOWN = "unknown"

	AUDIO_BITRATE_AUTO   = 0
	AUDIO_BITRATE_LOW    = 64
	AUDIO_BITRATE_MEDIUM = 128
	AUDIO_BITRATE_HIGH   = 192

	DEFAULT_DESTINATION     = "./%title%.%format%"
	DEFAULT_DESTINATION_MP3 = "./%title%.mp3"
)
