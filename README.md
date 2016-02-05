# Ward

Ward is a secure single-file password manager.

Ward stores your passwords in an encrypted file which you manage with a single master password. You can keep track of multiple complex passwords without having to remember any of them.

# Installation

[Download the zero-install binary on the releases page.](https://github.com/schmich/ward/releases)

# Usage

    Usage: ward [OPTIONS] COMMAND [arg...]
    
    Secure password manager - https://github.com/schmich/ward
    
    Options:
      -v, --version    Show the version and exit
    
    Commands:
      init         Create a new credential database.
      add          Add a new credential.
      copy         Copy a password to the clipboard.
      edit         Edit an existing credential.
      del          Delete a stored credential.
      show         Show a stored credential.
      import       Import JSON-formatted credentials.
      export       Export JSON-formatted credentials.
      master       Update master password.
  
    Run 'ward COMMAND --help' for more information on a command.

Create a new credential database:

    > ward init
    Creating new credential database.
    Master password:
    Master password (confirm):
    ✓ Credential database created at C:\Users\schmich\.ward.

Link to an existing credential database (e.g. from Dropbox):

    > ward init --link C:\Users\schmich\Dropbox\.ward
    ✓ Linked to existing database C:\Users\schmich\.ward -> C:\Users\schmich\Dropbox\.ward.

Add a new credential:

    > ward add
    Master password:
    Login: foo@example.com
    Password:
    Password (confirm):
    Realm: linkedin.com
    Note: LinkedIn account
    ✓ Credential added. Password copied to the clipboard.

Add a new credential with a generated password:

    > ward add --gen --length 15 --min-upper 1 --min-lower 1 --min-digit 1 --min-symbol 1 --no-similar
    Master password:
    Login: quux@example.com
    Realm: twitter.com
    Note: Twitter account
    ✓ Credential added. Generated password copied to the clipboard.

Copy an existing password:

    > ward copy linked
    Master password:
    ✓ Password for foo@example.com:linkedin.com copied to the clipboard.

Export credentials as JSON:

    > ward export
    Master password:
    [
      {
        "login": "foo@example.com",
        "password": "bH`-uKY~A1YG5T$SqNYN8pw,j!Xa\\Gsy41f|",
        "realm": "linkedin.com",
        "note": "LinkedIn account"
      }
    ]

Import JSON credentials:

    > ward import credentials.json
    Master password:
    Importing 192 credentials.
    ✓ Imported credentials from credentials.json.

# License

Copyright &copy; 2016 Chris Schmich<br>
MIT License. See [LICENSE](LICENSE) for details.
