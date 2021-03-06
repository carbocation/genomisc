# Downloading UKBB bulk data (cardiac MRI)
In brief, see [this instruction page](http://biobank.ndph.ox.ac.uk/showcase/instruct/bulk.html). However, these instructions are out of date (give the wrong argument order, etc). See updated instructions below.

Prepare permissions
1. Make sure that you have created a `.ukbkey` file containing the application ID on line 1 and the private key on line 2 (directly downloadable as an attachment from the email that you received from the UKBB). This file should not be readable by anyone without proper UKBB permissions, so consider setting this to be user-readable only.

Download data
1. Download the encrypted file (`ukb21481.enc`) and decrypt it to the encoded file (`ukb21481.enc_ukb`) by running `ukbunpack ukb21481.enc`
1. Extract any non-bulk data to TSV, e.g.: `ukbconv ukb21481.enc_ukb txt`
1. Extract the list of all samples with the field of interest. 
  * Attempt #1: 20208 is Heart MRI Long Axis `ukbconv ukb21481.enc_ukb bulk -s20208`
  * Atttempt #2: Try to get all MRI fields at once. `ukbconv ukb21481.enc_ukb bulk -ifields.list`
1. Inspect: `wc -l ukb21481.bulk` and you can see that there is one entry per person per field for whom this data exists
1. You cannot download more than 1,000 samples' bulk files at a time. So, iteratively do it:
    * For now, just take 50 
      * `head -n 50 ukb21481.bulk > heart.50`
      * `ukbfetch -bheart.50` *(Note: no space between `-b` and `heart.50`)*
    * There is also the `downloader.exe` tool that I created but it may be worse.
    * Try splitting and iterating. (This should be done in its own folder since the split files have no prefix nor suffix). *NOTE*: This is slower than the `downloader.exe` approach.
      * `split -l 1000 ukb21481.bulk`
      * `/bin/ls ./ | xargs -I {} ukbfetch -b{}`
    * When downloading, don't duplicate prior work:
      * `comm -23 newfile.bulk oldfile.bulk > new-samples-only.bulk`
      * This shows lines only present in the left file but not the right file. Useful if you have the bulk file from previous downloads.