package resource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

// priceAttributes returns the attributes of each data row of a price lookup fixture.
func priceAttributes(t *testing.T, fixture string) []json.RawMessage {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "testdata", "fixtures", "prices", fixture))
	require.NoError(t, err)
	var doc struct {
		Data []struct {
			Attributes json.RawMessage `json:"attributes"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &doc))
	attrs := make([]json.RawMessage, len(doc.Data))
	for i, d := range doc.Data {
		attrs[i] = d.Attributes
	}
	return attrs
}

func TestVoiceInPriceUnmarshal(t *testing.T) {
	attrs := priceAttributes(t, "voice_in/lookup/200.json")
	require.Len(t, attrs, 2)

	var priced, rejected VoiceInPrice
	require.NoError(t, json.Unmarshal(attrs[0], &priced))
	require.NoError(t, json.Unmarshal(attrs[1], &rejected))

	assert.Equal(t, "1800", priced.Destination)
	assert.Equal(t, "United States", *priced.Country)
	assert.Nil(t, priced.SrcCountryName)
	assert.Equal(t, "cellular", priced.SourceType)
	assert.Equal(t, "0.025", *priced.Rate)
	assert.Equal(t, "0.0", *priced.ConnectionFee)
	assert.Equal(t, 60, priced.InitialInterval)
	assert.Equal(t, 60, priced.NextInterval)
	assert.False(t, priced.RejectCalls)

	assert.Equal(t, "United Kingdom", *rejected.SrcCountryName)
	assert.Nil(t, rejected.Rate)
	assert.Nil(t, rejected.ConnectionFee)
	assert.True(t, rejected.RejectCalls)
}

func TestVoiceOutPriceUnmarshal(t *testing.T) {
	attrs := priceAttributes(t, "voice_out/lookup/200.json")
	require.Len(t, attrs, 2)

	var scheduled, plain VoiceOutPrice
	require.NoError(t, json.Unmarshal(attrs[0], &scheduled))
	require.NoError(t, json.Unmarshal(attrs[1], &plain))

	assert.Equal(t, enums.VoiceOutPriceCategoryInternational, scheduled.Category)
	assert.Equal(t, "Germany", *scheduled.Country)
	assert.Equal(t, "Fixed", *scheduled.Network)
	assert.Equal(t, "4930", scheduled.Prefix)
	assert.Equal(t, "0.0045", *scheduled.Rate)
	assert.Equal(t, "0.0", *scheduled.ConnectionFee)
	assert.Equal(t, 1, scheduled.InitialIncrement)
	assert.Equal(t, 1, scheduled.RecurringIncrement)
	assert.Equal(t, "0.005", *scheduled.NextRate)
	assert.Equal(t, "0.0", *scheduled.NextConnectionFee)
	assert.Equal(t, 30, *scheduled.NextInitialIncrement)
	assert.Equal(t, 6, *scheduled.NextRecurringIncrement)
	require.NotNil(t, scheduled.EffectiveAt)
	assert.True(t, scheduled.EffectiveAt.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "Increase and Billing increment change", *scheduled.Comment)
	assert.False(t, scheduled.RejectCalls)

	assert.Equal(t, enums.VoiceOutPriceCategoryOriginBased, plain.Category)
	assert.Nil(t, plain.NextRate)
	assert.Nil(t, plain.NextConnectionFee)
	assert.Nil(t, plain.NextInitialIncrement)
	assert.Nil(t, plain.NextRecurringIncrement)
	assert.Nil(t, plain.EffectiveAt)
	assert.Nil(t, plain.Comment)
}

func TestSMSOutPriceUnmarshal(t *testing.T) {
	attrs := priceAttributes(t, "sms_out/lookup/200.json")
	require.Len(t, attrs, 2)

	var priced, rejected SMSOutPrice
	require.NoError(t, json.Unmarshal(attrs[0], &priced))
	require.NoError(t, json.Unmarshal(attrs[1], &rejected))

	assert.Nil(t, priced.SourcePrefix)
	assert.Equal(t, "49151", *priced.DestinationPrefix)
	assert.Equal(t, "Germany", *priced.Country)
	assert.Equal(t, 262, *priced.MCC)
	assert.Equal(t, "Germany", *priced.MCCName)
	assert.Equal(t, 1, *priced.MNC)
	assert.Equal(t, "Telekom Deutschland", *priced.MNCName)
	assert.False(t, priced.RejectSMS)
	assert.Equal(t, "0.075000", *priced.Price)

	assert.True(t, rejected.RejectSMS)
	assert.Nil(t, rejected.Price)
}

func TestSMSInPriceUnmarshal(t *testing.T) {
	attrs := priceAttributes(t, "sms_in/lookup/200.json")
	require.Len(t, attrs, 1)

	var price SMSInPrice
	require.NoError(t, json.Unmarshal(attrs[0], &price))

	assert.Nil(t, price.SourcePrefix)
	assert.Equal(t, 1, price.SourceNumberLengthMin)
	assert.Equal(t, 15, price.SourceNumberLengthMax)
	assert.Equal(t, "1", *price.DestinationPrefix)
	assert.Equal(t, 11, price.DestinationNumberLengthMin)
	assert.Equal(t, 11, price.DestinationNumberLengthMax)
	assert.Equal(t, "United States", *price.Country)
	assert.Equal(t, "0.005000", price.Price)
}
