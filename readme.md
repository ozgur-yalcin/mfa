# mfa

mfa *("multi-factor authentication")* is a command line tool for generating and validating one-time password.

<!-- TOC -->

  * [Usage](#usage)
  * [Examples](#examples)
    + [Generate code](#generate-code)
    + [Add account](#add-account)
    + [List account](#list-account)
    + [Remove accounts](#remove-accounts)
    + [Update account](#update-account)
  * [License](#license)

<!-- /TOC -->

**Description**:

* An easy-to-use substitute for 2FA apps like TOTP Google authenticator.
* Supports the OATH algorithms, such as TOTP and HOTP.
* No need for network connection.
* No need for phone.

## Usage

```
mfa [command] [flags] [args]
```

```
Available Commands:
  add         Add account and its secret key
  completion  Generate the autocompletion script for the specified shell
  generate    Generate one-time password from secret key
  help        Help about any command
  list        List all added accounts and password code
  remove      Remove account and its secret key
  update      Add account and its secret key
  version     show version
```

```
mfa generate [flags] <secret key>
mfa add [flags] <account name> <secret key>
mfa remove <account name> [user name]
mfa update [flags] <account name> <secret key>
mfa list [account name]
```

Commonly used flags

```
Flags:
  -b, --base32         use base32 encoding of KEY instead of hex (default true)
  -c, --counter int    used for HOTP, A counter C, which counts the number of iterations
  -e, --epoch int      used for TOTP, epoch (T0) which is the Unix time from which to start counting time steps
  -H, --hash string    A cryptographic hash method H (SHA1, SHA256, SHA512) (default "SHA1")
  -h, --help           help for generate
  -i, --interval int   used for TOTP, an interval (Tx) which will be used to calculate the value of the counter CT (default 30)
  -l, --length int     A HOTP value length d (default 6)
  -m, --mode string    use time-variant TOTP mode or use event-based HOTP mode (default "totp")
```

## Examples

### Generate code

Generate a **time-based** one-time password but do not save the secret key

```
mfa generate ADOO3MCCCVO5AVD6
```

Generate a **counter-based** one-time password with counter 1

```
mfa generate -m hotp -c 1 ADOO3MCCCVO5AVD6
```

### Add account

Add an account named GitHub

```
mfa add GitHub ADOO3MCCCVO5AVD6
```

Add an account, the account name is GitHub, the user name is ozgur-yalcin

```
mfa add GitHub:ozgur-yalcin ADOO3MCCCVO5AVD6
```

### List account

List all accounts

```shell
mfa list 
```

List all accounts named GitHub

```
mfa list GitHub
```

List accounts whose account name is GitHub and whose username is ozgur-yalcin

```
mfa list GitHub:ozgur-yalcin
```

List accounts whose account name is GitHub and whose username is ozgur-yalcin

```
mfa list GitHub ozgur-yalcin
```

### Remove accounts

Remove all accounts named GitHub

```
mfa remove GitHub
```

Delete accounts  whose account name is GitHub and whose username is ozgur-yalcin

```
mfa remove GitHub ozgur-yalcin
```

### Update account

Update the secret key of accounts which account name is GitHub

```
mfa update GitHub 5BRSSSBJUWBQBOXE
```

Update the secret key of accounts which account name is GitHub and the username is ozgur-yalcin

```
mfa update GitHub:ozgur-yalcin 5BRSSSBJUWBQBOXE
```

## License

MIT License, see [license.md](license.md).
