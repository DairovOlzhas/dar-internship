package discussion

import "errors"

var (
	ErrNotFound            = errors.New("Discussion not found.")
	ErrMessageNotFound     = errors.New("Message not found.")
	ErrMessageCreation     = errors.New("Message creation error.")
	ErrInvalidMessage      = errors.New("Message invalid.")
	ErrFileCreation        = errors.New("File creation error.")
	ErrViolationCreation   = errors.New("Violation creation error.")
	ErrDiscussionCreation  = errors.New("Were not able to create dicussion.")
	ErrInvalidDiscussion   = errors.New("Discussion invalid.")
	ErrUpdate              = errors.New("Update error.")
	ErrNothingToUpdate     = errors.New("Nothing to update.")
	ErrNoRecipient         = errors.New("No recipient ID.")
	ErrConnClosed          = errors.New("Operating over closed connection.")
	ErrReadOwnMessage      = errors.New("Can't read own message.")
	ErrReadMessage         = errors.New("Message has already been read.")
	ErrParticipantExists   = errors.New("Participant already exists.")
	ErrParticipantNotFound = errors.New("Participant not found.")
	ErrParticipantUpdate   = errors.New("Unable to update participant info.")
	ErrNoPermission        = errors.New("No permission.")
)
