# export-mondo-transactions [![Build Status](https://travis-ci.org/Sam-Martin/export-mondo-transactions.svg?branch=master)](https://travis-ci.org/Sam-Martin/export-mondo-transactions)
This Go package allows you to create OFX files from your Mondo transaction history. It was created for and tested against YNAB, so there are no guarantees that the OFX files generated will be usable by other apps.

# Usage  

0. Install Go
1. Create a [client from Mondo's website](https://developers.getmondo.co.uk/apps/home) (redirect url doesn't matter, use google.com if you like).
2. Identify the `client_id` and `client_secret` from the newly created client
3. Download the [latest release](https://github.com/Sam-Martin/export-mondo-transactions/releases/latest)
4. Run the downloaded executable andenter your `client_id` and `client_secret` when prompted
5. Follow the instructions in the browser that opens to `localhost:8080`

# How it works
This package runs its own webserver to make the oAuth process easier (possible really!) and upon receiving authorisation (by you following the link emailed to you by Mondo) will automatically enumerate your transactions and write an OFX file for you.
