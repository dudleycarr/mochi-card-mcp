package mochi

import "strings"

// sideSeparator is the line Mochi uses to separate the two sides of a card.
const sideSeparator = "\n---\n"

// JoinSides combines the front and back of a card into Mochi's single-field
// Markdown representation, where the two sides are separated by a line
// containing only "---". If back is empty, front is returned unchanged.
func JoinSides(front, back string) string {
	if back == "" {
		return front
	}
	return front + sideSeparator + back
}

// SplitSides splits Mochi card content into its front and back halves on the
// first "---" separator line. If there is no separator, the whole content is
// returned as the front and back is empty.
func SplitSides(content string) (front, back string) {
	if i := strings.Index(content, sideSeparator); i >= 0 {
		return content[:i], content[i+len(sideSeparator):]
	}
	return content, ""
}
