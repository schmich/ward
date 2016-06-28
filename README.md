# Ward

Ward is a secure, cross-platform, zero-install, single-file password manager.

Ward stores your passwords in an encrypted file which you manage with a single master password. You can keep track of multiple complex passwords without having to remember any of them.

# Install

- [Download the zero-install binary](https://github.com/schmich/ward/releases) to a directory on your `PATH`; or
- `go get -u github.com/schmich/ward/... && go install github.com/schmich/ward/...`

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
      qr           Print password formatted as a QR code.
      import       Import JSON-formatted credentials.
      export       Export JSON-formatted credentials.
      list         Print a table-formatted list of credentials.
      master       Update master password.

    Run 'ward COMMAND --help' for more information on a command.

Create a new credential database:

    > ward init
    Creating new credential database.
    Master password:
    Master password (confirm):
    ✓ Credential database created at C:\Users\schmich\.ward.

Link to an existing credential database. This requires administrator privileges on Windows. A symlink will be created to the specified file:

    > ward init --link C:\Users\schmich\Dropbox\.ward
    ✓ Linked to existing database C:\Users\schmich\.ward -> C:\Users\schmich\Dropbox\.ward.

Add a new credential:

    > ward add
    Master password:
    Login: fizz@buzz.com
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

Copy an existing password with partial string matching:

    > ward copy linked
    Master password:
    ✓ Password for fizz@buzz.com@linkedin.com copied to the clipboard.

Export credentials as JSON:

    > ward export
    Master password:
    [
      {
        "login": "fizz@buzz.com",
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

The Ward database is stored at `~/.ward`. This can be overridden with the `WARDFILE` environment variable, e.g. in `.bashrc`:

    export WARDFILE=~/dotfiles/ward

# Password Generator

Ward comes with a constraint-solving password generator that you can use when adding a new credential (`ward add --gen`). You can control length, character requirements, and exclusions:

    > ward add --help

    Usage: ward add [--login] [--realm] [--note] [--no-copy] [--gen [--length] [--min-length] [--max-length] [--no-upper] [--no-lower] [--no-digit] [--no-symbol] [--no-similar] [--min-upper] [--max-upper] [--min-lower] [--max-lower] [--min-digit] [--max-digit] [--min-symbol] [--max-symbol] [--exclude]]

    Options:
      --login=""           Login for credential, e.g. username or email.
      --realm=""           Realm for credential, e.g. website or WiFi AP name.
      --note=""            Note for credential.
      --no-copy=false      Do not copy password to the clipboard.
      --gen=false          Generate a password.
      --length=0           Password length.
      --min-length=30      Minimum length password.
      --max-length=40      Maximum length password.
      --no-upper=false     Exclude uppercase characters from password.
      --no-lower=false     Exclude lowercase characters from password.
      --no-digit=false     Exclude digit characters from password.
      --no-symbol=false    Exclude symbol characters from password.
      --no-similar=false   Exclude similar characters from password: 5SB8|1IiLl0Oo.
      --min-upper=0        Minimum number of uppercase characters in password.
      --max-upper=-1       Maximum number of uppercase characters in password.
      --min-lower=0        Minimum number of lowercase characters in password.
      --max-lower=-1       Maximum number of lowercase characters in password.
      --min-digit=0        Minimum number of digit characters in password.
      --max-digit=-1       Maximum number of digit characters in password.
      --min-symbol=0       Minimum number of symbol characters in password.
      --max-symbol=-1      Maximum number of symbol characters in password.
      --exclude=""         Exclude specific characters from password.

# License

Copyright &copy; 2016 Chris Schmich<br>
MIT License. See [LICENSE](LICENSE) for details.
