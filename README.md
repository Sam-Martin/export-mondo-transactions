# export-monzo-transactions [![Build Status](https://travis-ci.org/Sam-Martin/export-mondo-transactions.svg?branch=master)](https://travis-ci.org/Sam-Martin/export-mondo-transactions) [![Build status](https://ci.appveyor.com/api/projects/status/78r5jpxk7lacl28s/branch/master?svg=true)](https://ci.appveyor.com/project/Sam-Martin/export-mondo-transactions-c0aws/branch/master)

This Go package allows you to create OFX files from your Monzo transaction history. It was created for and tested against YNAB, so there are no guarantees that the OFX files generated will be usable by other apps.  
![Screenshot](https://cloud.githubusercontent.com/assets/803607/15273152/519c2698-1a88-11e6-9cf9-7e8314da1b3b.png)  

# Usage  

1. Create a [client from Monzo's website](https://developers.monzo.com/) (Redirect URL **MUST** be `http://localhost:8080/getTransactions/`).
2. Identify the `client_id` and `client_secret` from the newly created client
3. Download the [latest release](https://github.com/Sam-Martin/export-mondo-transactions/releases/latest)
4. Run the downloaded executable and enter your `client_id` and `client_secret` when prompted
5. Follow the instructions in the browser that opens to `localhost:8080`

# How it works
This package runs its own webserver to make the oAuth process easier (possible really!) and upon receiving authorisation (by you following the link emailed to you by Monzo) will automatically enumerate your transactions and write an OFX file for you.

# Debugging
Run the executable with `\export-monzo-transactions.exe --logLevel debug`

# Caveats & Contributing
This was my first experience coding in Go and a result it is very [WET](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself), it may improve in time depending on my schedule.  
If you want to take a crack at improving it in any way, just reach out to me via email/[Twitter](https://twitter.com/samjackmartin) and send a PR!
