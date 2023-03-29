# DALL-E-go

golang version DALL-E API

## HOWTO

```golang
// create client
client := dalle.NewClient(OPENAI_API_KEY)

// request image by prompt
response, err := client.Request("A cute yeti")

// fetch image binary data
for _, data := range response.Data {
  binaryData, _ := data.Binary()
}
```