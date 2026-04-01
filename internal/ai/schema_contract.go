package ai

const jotformSchemaContract = `
A valid Jotform form schema is a JSON object with this structure:

{
  "questions": {
    "<order>": {
      "type": "<field_type>",
      "text": "<label>",
      "name": "<snake_case_name>",
      "order": "<integer_as_string>",
      "required": "No" | "Yes"
    }
  },
  "properties": {
    "title": "<form title>"
  }
}

Supported field types: control_textbox, control_textarea, control_email,
control_phone, control_number, control_radio, control_checkbox,
control_dropdown, control_fileupload, control_date, control_address,
control_head (section header), control_divider, control_button (submit).

For radio/checkbox/dropdown, add an "options" field with pipe-separated values:
  "options": "Option A|Option B|Option C"

The first question (order "1") should always be a control_head with the form title.
The last question should be a control_button with text "Submit".

Return ONLY the JSON object, no markdown fences, no explanation.
`
