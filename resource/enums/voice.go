package enums

// CliFormat defines the CLI (Caller Line Identification) format for voice trunks.
type CliFormat string

const (
	CliFormatRaw   CliFormat = "raw"
	CliFormatE164  CliFormat = "e164"
	CliFormatLocal CliFormat = "local"
)

// Points of presence a voice in trunk can be pinned to (VoiceInTrunk.Pop).
// A nil Pop lets DIDWW choose.
const (
	PopNYC = "nyc"
	PopFRA = "fra"
	PopLAC = "lac"
	PopMIA = "mia"
	PopSG  = "sg"
	PopHK  = "hk"
	PopAMS = "ams"
)

// OnCliMismatchAction defines the action taken when CLI does not match any DID.
type OnCliMismatchAction string

const (
	OnCliMismatchActionSendOriginalCli OnCliMismatchAction = "send_original_cli"
	OnCliMismatchActionRejectCall      OnCliMismatchAction = "reject_call"
	// OnCliMismatchActionReplaceCli requires account configuration. Contact DIDWW support to enable.
	OnCliMismatchActionReplaceCli OnCliMismatchAction = "replace_cli"
	// OnCliMismatchActionRandomizeCli requires account configuration. Contact DIDWW support to enable.
	OnCliMismatchActionRandomizeCli OnCliMismatchAction = "randomize_cli"
)

// MediaEncryptionMode defines the media encryption mode for voice trunks.
type MediaEncryptionMode string

const (
	MediaEncryptionModeDisabled MediaEncryptionMode = "disabled"
	MediaEncryptionModeSrtpSdes MediaEncryptionMode = "srtp_sdes"
	MediaEncryptionModeSrtpDtls MediaEncryptionMode = "srtp_dtls"
	MediaEncryptionModeZrtp     MediaEncryptionMode = "zrtp"
)

// DefaultDstAction defines the default destination action for voice-out trunks.
type DefaultDstAction string

const (
	DefaultDstActionAllowAll  DefaultDstAction = "allow_all"
	DefaultDstActionRejectAll DefaultDstAction = "reject_all"
)

// VoiceOutTrunkStatus defines the status of a voice-out trunk.
type VoiceOutTrunkStatus string

const (
	VoiceOutTrunkStatusActive  VoiceOutTrunkStatus = "active"
	VoiceOutTrunkStatusBlocked VoiceOutTrunkStatus = "blocked"
)

// StirShakenMode defines the STIR/SHAKEN attestation mode for voice-in trunks.
type StirShakenMode string

const (
	StirShakenModeDisabled    StirShakenMode = "disabled"
	StirShakenModeOriginal    StirShakenMode = "original"
	StirShakenModePai         StirShakenMode = "pai"
	StirShakenModeOriginalPai StirShakenMode = "original_pai"
	StirShakenModeVerstat     StirShakenMode = "verstat"
)

// TransportProtocol defines the SIP transport protocol.
type TransportProtocol int

const (
	TransportProtocolUDP TransportProtocol = 1
	TransportProtocolTCP TransportProtocol = 2
	TransportProtocolTLS TransportProtocol = 3
)

// RxDtmfFormat defines the receive DTMF format for SIP trunks.
type RxDtmfFormat int

const (
	RxDtmfFormatRFC2833          RxDtmfFormat = 1
	RxDtmfFormatSIPInfo          RxDtmfFormat = 2
	RxDtmfFormatRFC2833OrSIPInfo RxDtmfFormat = 3
)

// TxDtmfFormat defines the transmit DTMF format for SIP trunks.
type TxDtmfFormat int

const (
	TxDtmfFormatDisabled     TxDtmfFormat = 1
	TxDtmfFormatRFC2833      TxDtmfFormat = 2
	TxDtmfFormatSIPInfoRelay TxDtmfFormat = 3
	TxDtmfFormatSIPInfoDtmf  TxDtmfFormat = 4
)

// SstRefreshMethod defines the SIP session timer refresh method.
type SstRefreshMethod int

const (
	SstRefreshMethodInvite               SstRefreshMethod = 1
	SstRefreshMethodUpdate               SstRefreshMethod = 2
	SstRefreshMethodUpdateFallbackInvite SstRefreshMethod = 3
)

// Codec is a codec of an inbound SIP trunk (SIPConfiguration.Codecs).
// Voice OUT trunks have their own list, VoiceOutCodec.
type Codec string

const (
	CodecTelephoneEvent Codec = "telephone_event"
	CodecG723           Codec = "g723"
	CodecG729           Codec = "g729"
	CodecPCMU           Codec = "pcmu"
	CodecPCMA           Codec = "pcma"
	CodecSpeex          Codec = "speex"
	CodecGSM            Codec = "gsm"
	CodecG726_32        Codec = "g726_32"
	CodecG721           Codec = "g721"
	CodecG726_24        Codec = "g726_24"
	CodecG726_40        Codec = "g726_40"
	CodecG726_16        Codec = "g726_16"
	CodecL16            Codec = "l16"
	CodecAMR            Codec = "amr"
	CodecAMRWB          Codec = "amr_wb"
)

// VoiceOutCodec is a codec of a Voice OUT trunk (VoiceOutTrunk.Codecs).
type VoiceOutCodec string

const (
	VoiceOutCodecTelephoneEvent VoiceOutCodec = "telephone_event"
	VoiceOutCodecG723           VoiceOutCodec = "g723"
	VoiceOutCodecG729           VoiceOutCodec = "g729"
	VoiceOutCodecPCMU           VoiceOutCodec = "pcmu"
	VoiceOutCodecPCMA           VoiceOutCodec = "pcma"
	VoiceOutCodecGSM            VoiceOutCodec = "gsm"
	VoiceOutCodecG722           VoiceOutCodec = "g722"
	VoiceOutCodecOpus           VoiceOutCodec = "opus"
	VoiceOutCodecAMR            VoiceOutCodec = "amr"
	VoiceOutCodecAMRWB          VoiceOutCodec = "amr_wb"
)

// IceMode defines ICE (Interactive Connectivity Establishment) on the calls of a trunk.
type IceMode string

const (
	IceModeDisabled IceMode = "disabled"
	IceModeAccept   IceMode = "accept"
	IceModeInitiate IceMode = "initiate"
)

// RtcpMuxMode defines RTP and RTCP multiplexed on one port on the calls of a trunk.
type RtcpMuxMode string

const (
	RtcpMuxModeDisabled RtcpMuxMode = "disabled"
	RtcpMuxModeAccept   RtcpMuxMode = "accept"
	RtcpMuxModeInitiate RtcpMuxMode = "initiate"
)

// RtcpFeedbackMode defines RTCP feedback (AVPF) on the calls of a trunk.
type RtcpFeedbackMode string

const (
	RtcpFeedbackModeDisabled RtcpFeedbackMode = "disabled"
	RtcpFeedbackModeAccept   RtcpFeedbackMode = "accept"
	RtcpFeedbackModeInitiate RtcpFeedbackMode = "initiate"
)
