package data

import "time"

const RelationOwner = "owner"
const RelationRead = "read"

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Authentication struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Login  string   `json:"login"`
	Rights []string `json:"rights"`
}

type User struct {
	Id         string `json:"id"`
	Login      string `json:"login"`
	FullName   string `json:"fullName"`
	ProvidedId string `json:"providedId"`
}

type Journal struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type JournalItem struct {
	Id          string    `json:"id"`
	JournalID   string    `json:"journalId"`
	Date        string    `json:"date"`
	Comments    string    `json:"comments"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Attachment struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}
type CompleteJournalItem struct {
	Id          string       `json:"id"`
	JournalID   string       `json:"journalId"`
	Date        string       `json:"date"`
	Comments    string       `json:"comments"`
	LastUpdated time.Time    `json:"lastUpdated"`
	Attachments []Attachment `json:"attachments"`
}

type CompleteJournal struct {
	Journal Journal `json:"journal"`
	Owner   bool    `json:"owner"`
	User    User    `json:"user"`
}

type LoginResponse struct {
	User     User              `json:"user"`
	Journals []CompleteJournal `json:"journals"`
}

type SearchData struct {
	Search string `json:"search"`
}

type IdsHolder struct {
	Ids []string `json:"ids"`
}

type TitleHolder struct {
	AttachmentID string `json:"attachmentId"`
	Title        string `json:"title"`
}

type Titles struct {
	Titles []TitleHolder `json:"titles"`
}

type StringHolder struct {
	Val string `json:"val"`
}

type JournalUsersRequest struct {
	JournalID string   `json:"journalId"`
	UserIDs   []string `json:"userIds"`
}

type JournalUpdateData struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
