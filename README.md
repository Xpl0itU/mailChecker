### Go IMAP Email Checker
This Go program checks emails using the IMAP protocol and performs actions based on predefined filters. It is designed to move emails to different folders depending on the filter criteria.

## Prerequisites
Before using the program, ensure you meet the following requirements:

Binaries: You can obtain pre-built binaries from the Releases section of this repository.

IMAP Email Server: Access to an IMAP email server is required for checking and managing emails.

## Usage
Download the appropriate binary for your operating system from the Releases page.

Create a .env file in the same directory as the binary with the following environment variables:

```
SERVER=your_imap_server:port
EMAIL=your_email_address
PASSWORD=your_email_password
MAIL_OK_FOLDER=MailOkFolder
MAIL_FAILED_FOLDER=MailFailedFolder
```
Replace the placeholders with your IMAP server details and folder names.

Create a filters.json file with your email filtering rules. Example filters.json:

```json
[
    {
        "mail": "sender@example.com",
        "subject": "Test",
        "fail_if_found": true,
        "hour_threshold": 24,
        "comment": "",
        "fail_if_not_found": true
    }
]
```
Define multiple filters as needed.

Run the binary in your terminal:

```shell
./mailChecker
```
The program will connect to your IMAP server, apply the defined filters to your emails, and move them to the specified folders based on the filter criteria.

Contributing
Contributions to this project are welcome. Please submit issues or pull requests to improve the program. Your contributions are highly appreciated!

License
This project is licensed under the GNU General Public License v3.0 (GPL-3.0). You can find the full license text in the LICENSE file.
