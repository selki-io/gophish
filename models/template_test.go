package models

import (
	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestTemplateLanguageValidation(c *check.C) {
	// Test valid language codes
	validLangs := []Language{LangEN, LangES, LangPT, ""}
	for _, lang := range validLangs {
		t := Template{
			Name:   "Test Template",
			Text:   "Test content",
			UserId: 1,
			Lang:   lang,
		}
		err := t.Validate()
		c.Assert(err, check.Equals, nil)
	}

	// Test invalid language code
	t := Template{
		Name:   "Test Template",
		Text:   "Test content",
		UserId: 1,
		Lang:   "fr", // Invalid language
	}
	err := t.Validate()
	c.Assert(err, check.Equals, ErrTemplateInvalidLanguage)
}

func (s *ModelsSuite) TestTemplateLanguagePersistence(c *check.C) {
	// Create template with language
	t := Template{
		Name:   "Spanish Template",
		Text:   "Contenido de prueba",
		UserId: 1,
		Lang:   LangES,
	}

	// Post template
	err := PostTemplate(&t)
	c.Assert(err, check.Equals, nil)

	// Retrieve template and verify language
	retrieved, err := GetTemplate(t.Id, 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(retrieved.Lang, check.Equals, LangES)

	// Clean up
	err = DeleteTemplate(t.Id, 1)
	c.Assert(err, check.Equals, nil)
}
