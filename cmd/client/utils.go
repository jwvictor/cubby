package main

import (
	client2 "github.com/jwvictor/cubby/pkg/client"
	"github.com/spf13/viper"
	"regexp"
	"strings"
	"bufio"
	"os"
)

func getClient() *client2.CubbyClient {
	//log.Printf("Connecting user %s to %s:%d\n", viper.GetString(CfgUserEmail), viper.GetString(CfgHost), viper.GetInt(CfgPort))
	client := client2.NewCubbyClient(viper.GetString(CfgHost),
		viper.GetInt(CfgPort),
		viper.GetString(CfgUserEmail),
		viper.GetString(CfgUserPassword))
	return client
}

func extractTags(comment string) []string {
	re := regexp.MustCompile("#\\S+")
	tags := re.FindAllString(comment, -1)

	// Strip #s
	for i, tag := range tags {
		tags[i] = strings.ToLower(strings.TrimLeft(tag, "#"))
	}

	return tags
}

func deduplicateTags(tagArray []string) []string {
	tags := map[string]bool{}

	// Add all flags from #tag
	for _, tag := range tagArray {
		tags[tag] = true
	}

	// Deduplicate
	uniqueTags := []string{}
	for tag := range tags {
		uniqueTags = append(uniqueTags, tag)
	}

	return uniqueTags

}

func slurpStdin() []byte {

  var out []byte
  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
    //fmt.Println(scanner.Text())
    out = append(out, scanner.Bytes()...)
  }

  if scanner.Err() != nil {
    return nil
  }
  return out
}
