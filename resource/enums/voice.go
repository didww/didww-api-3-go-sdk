package enums

// CliFormat defines the CLI (Caller Line Identification) format for voice trunks.
type CliFormat string

const (
	CliFormatRaw   CliFormat = "raw"
	CliFormatE164  CliFormat = "e164"
	CliFormatLocal CliFormat = "local"
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

// Codec defines audio codec identifiers for voice trunks.
type Codec int

const (
	CodecTelephoneEvent Codec = 6
	CodecG723           Codec = 7
	CodecG729           Codec = 8
	CodecPCMU           Codec = 9
	CodecPCMA           Codec = 10
	CodecSpeex          Codec = 12
	CodecGSM            Codec = 13
	CodecG726_32        Codec = 14
	CodecG721           Codec = 15
	CodecG726_24        Codec = 16
	CodecG726_40        Codec = 17
	CodecG726_16        Codec = 18
	CodecL16            Codec = 19
)
