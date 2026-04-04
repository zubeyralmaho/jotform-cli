package templates

// Template represents a curated form starter template.
type Template struct {
	Name        string
	Description string
	Category    string
	Schema      map[string]interface{}
}

// Builtin returns all curated starter templates.
func Builtin() []Template {
	return []Template{
		contactForm(),
		feedbackForm(),
		rsvpForm(),
		orderForm(),
		registrationForm(),
		surveyForm(),
		bugReportForm(),
		jobApplicationForm(),
	}
}

// Get returns a template by name, or nil if not found.
func Get(name string) *Template {
	for _, t := range Builtin() {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func contactForm() Template {
	return Template{
		Name:        "contact",
		Description: "Simple contact form with name, email, and message",
		Category:    "General",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Contact Us",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Contact Us",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Full Name",
					"name":     "fullName",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_email",
					"text":     "Email Address",
					"name":     "email",
					"order":    "3",
					"required": "Yes",
				},
				"4": map[string]interface{}{
					"type":  "control_textbox",
					"text":  "Subject",
					"name":  "subject",
					"order": "4",
				},
				"5": map[string]interface{}{
					"type":     "control_textarea",
					"text":     "Message",
					"name":     "message",
					"order":    "5",
					"required": "Yes",
				},
				"6": map[string]interface{}{
					"type":  "control_button",
					"text":  "Send Message",
					"order": "6",
				},
			},
		},
	}
}

func feedbackForm() Template {
	return Template{
		Name:        "feedback",
		Description: "Customer feedback form with rating and comments",
		Category:    "General",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Feedback Form",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "We'd Love Your Feedback",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Your Name",
					"name":     "name",
					"order":    "2",
					"required": "No",
				},
				"3": map[string]interface{}{
					"type":  "control_email",
					"text":  "Email",
					"name":  "email",
					"order": "3",
				},
				"4": map[string]interface{}{
					"type":     "control_rating",
					"text":     "Overall Rating",
					"name":     "rating",
					"order":    "4",
					"required": "Yes",
				},
				"5": map[string]interface{}{
					"type":  "control_dropdown",
					"text":  "What area is your feedback about?",
					"name":  "category",
					"order": "5",
					"options": "Product|Customer Service|Website|Pricing|Other",
				},
				"6": map[string]interface{}{
					"type":     "control_textarea",
					"text":     "Your Feedback",
					"name":     "feedback",
					"order":    "6",
					"required": "Yes",
				},
				"7": map[string]interface{}{
					"type":  "control_button",
					"text":  "Submit Feedback",
					"order": "7",
				},
			},
		},
	}
}

func rsvpForm() Template {
	return Template{
		Name:        "rsvp",
		Description: "Event RSVP form with attendance and guest count",
		Category:    "Events",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Event RSVP",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Event RSVP",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Your Name",
					"name":     "name",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_email",
					"text":     "Email Address",
					"name":     "email",
					"order":    "3",
					"required": "Yes",
				},
				"4": map[string]interface{}{
					"type":     "control_radio",
					"text":     "Will you attend?",
					"name":     "attendance",
					"order":    "4",
					"required": "Yes",
					"options":  "Yes, I'll be there|No, I can't make it|Maybe",
				},
				"5": map[string]interface{}{
					"type":    "control_spinner",
					"text":    "Number of Guests",
					"name":    "guests",
					"order":   "5",
				},
				"6": map[string]interface{}{
					"type":  "control_textarea",
					"text":  "Dietary Requirements or Notes",
					"name":  "notes",
					"order": "6",
				},
				"7": map[string]interface{}{
					"type":  "control_button",
					"text":  "Submit RSVP",
					"order": "7",
				},
			},
		},
	}
}

func orderForm() Template {
	return Template{
		Name:        "order",
		Description: "Simple product order form",
		Category:    "Business",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Order Form",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Place Your Order",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Full Name",
					"name":     "fullName",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_email",
					"text":     "Email",
					"name":     "email",
					"order":    "3",
					"required": "Yes",
				},
				"4": map[string]interface{}{
					"type":     "control_phone",
					"text":     "Phone Number",
					"name":     "phone",
					"order":    "4",
					"required": "Yes",
				},
				"5": map[string]interface{}{
					"type":     "control_dropdown",
					"text":     "Product",
					"name":     "product",
					"order":    "5",
					"required": "Yes",
					"options":  "Product A|Product B|Product C",
				},
				"6": map[string]interface{}{
					"type":  "control_spinner",
					"text":  "Quantity",
					"name":  "quantity",
					"order": "6",
				},
				"7": map[string]interface{}{
					"type":     "control_address",
					"text":     "Shipping Address",
					"name":     "address",
					"order":    "7",
					"required": "Yes",
				},
				"8": map[string]interface{}{
					"type":  "control_textarea",
					"text":  "Special Instructions",
					"name":  "instructions",
					"order": "8",
				},
				"9": map[string]interface{}{
					"type":  "control_button",
					"text":  "Place Order",
					"order": "9",
				},
			},
		},
	}
}

func registrationForm() Template {
	return Template{
		Name:        "registration",
		Description: "User registration form with account details",
		Category:    "General",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Registration Form",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Create Your Account",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Full Name",
					"name":     "fullName",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_email",
					"text":     "Email Address",
					"name":     "email",
					"order":    "3",
					"required": "Yes",
				},
				"4": map[string]interface{}{
					"type":     "control_phone",
					"text":     "Phone Number",
					"name":     "phone",
					"order":    "4",
					"required": "No",
				},
				"5": map[string]interface{}{
					"type":    "control_date",
					"text":    "Date of Birth",
					"name":    "dob",
					"order":   "5",
				},
				"6": map[string]interface{}{
					"type":  "control_address",
					"text":  "Address",
					"name":  "address",
					"order": "6",
				},
				"7": map[string]interface{}{
					"type":  "control_button",
					"text":  "Register",
					"order": "7",
				},
			},
		},
	}
}

func surveyForm() Template {
	return Template{
		Name:        "survey",
		Description: "General survey with multiple question types",
		Category:    "Research",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Survey",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Quick Survey",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_radio",
					"text":     "How satisfied are you with our service?",
					"name":     "satisfaction",
					"order":    "2",
					"required": "Yes",
					"options":  "Very Satisfied|Satisfied|Neutral|Dissatisfied|Very Dissatisfied",
				},
				"3": map[string]interface{}{
					"type":    "control_scale",
					"text":    "How likely are you to recommend us? (1-10)",
					"name":    "nps",
					"order":   "3",
				},
				"4": map[string]interface{}{
					"type":    "control_checkbox",
					"text":    "Which features do you use?",
					"name":    "features",
					"order":   "4",
					"options": "Forms|Reports|Submissions|API|Integrations",
				},
				"5": map[string]interface{}{
					"type":  "control_textarea",
					"text":  "Any additional comments?",
					"name":  "comments",
					"order": "5",
				},
				"6": map[string]interface{}{
					"type":  "control_button",
					"text":  "Submit Survey",
					"order": "6",
				},
			},
		},
	}
}

func bugReportForm() Template {
	return Template{
		Name:        "bug-report",
		Description: "Bug report form for software projects",
		Category:    "Engineering",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Bug Report",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Report a Bug",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_textbox",
					"text":     "Bug Title",
					"name":     "title",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_dropdown",
					"text":     "Severity",
					"name":     "severity",
					"order":    "3",
					"required": "Yes",
					"options":  "Critical|High|Medium|Low",
				},
				"4": map[string]interface{}{
					"type":     "control_dropdown",
					"text":     "Component",
					"name":     "component",
					"order":    "4",
					"options":  "Frontend|Backend|API|Database|Other",
				},
				"5": map[string]interface{}{
					"type":     "control_textarea",
					"text":     "Steps to Reproduce",
					"name":     "steps",
					"order":    "5",
					"required": "Yes",
				},
				"6": map[string]interface{}{
					"type":     "control_textarea",
					"text":     "Expected Behavior",
					"name":     "expected",
					"order":    "6",
					"required": "Yes",
				},
				"7": map[string]interface{}{
					"type":  "control_textarea",
					"text":  "Actual Behavior",
					"name":  "actual",
					"order": "7",
				},
				"8": map[string]interface{}{
					"type":  "control_fileupload",
					"text":  "Screenshots",
					"name":  "screenshots",
					"order": "8",
				},
				"9": map[string]interface{}{
					"type":  "control_button",
					"text":  "Submit Report",
					"order": "9",
				},
			},
		},
	}
}

func jobApplicationForm() Template {
	return Template{
		Name:        "job-application",
		Description: "Job application form with resume upload",
		Category:    "HR",
		Schema: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": "Job Application",
			},
			"questions": map[string]interface{}{
				"1": map[string]interface{}{
					"type":  "control_head",
					"text":  "Job Application",
					"order": "1",
				},
				"2": map[string]interface{}{
					"type":     "control_fullname",
					"text":     "Full Name",
					"name":     "fullName",
					"order":    "2",
					"required": "Yes",
				},
				"3": map[string]interface{}{
					"type":     "control_email",
					"text":     "Email Address",
					"name":     "email",
					"order":    "3",
					"required": "Yes",
				},
				"4": map[string]interface{}{
					"type":  "control_phone",
					"text":  "Phone Number",
					"name":  "phone",
					"order": "4",
				},
				"5": map[string]interface{}{
					"type":     "control_dropdown",
					"text":     "Position Applied For",
					"name":     "position",
					"order":    "5",
					"required": "Yes",
					"options":  "Software Engineer|Product Manager|Designer|Data Analyst|Other",
				},
				"6": map[string]interface{}{
					"type":     "control_fileupload",
					"text":     "Upload Resume",
					"name":     "resume",
					"order":    "6",
					"required": "Yes",
				},
				"7": map[string]interface{}{
					"type":  "control_textarea",
					"text":  "Cover Letter / Additional Notes",
					"name":  "coverLetter",
					"order": "7",
				},
				"8": map[string]interface{}{
					"type":  "control_textbox",
					"text":  "LinkedIn Profile URL",
					"name":  "linkedin",
					"order": "8",
				},
				"9": map[string]interface{}{
					"type":  "control_button",
					"text":  "Submit Application",
					"order": "9",
				},
			},
		},
	}
}
