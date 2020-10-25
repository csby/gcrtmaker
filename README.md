# gcrtmaker
Tool of making certificate files for windows

## openssl
###### check the contents of the CRL
`openssl crl -noout -text -in cr.crl`

###### verify the CRL
`openssl crl -noout -CAfile ca.crt -in cr.crl`

###### change the CRL from pem to der format
`openssl crl -in cr.crl -outform DER -out cr.der`

###### change the CRL from der to pem format
`openssl crl -inform DER -in cr.der -outform PEM -out cr.pem`

###### create Diffie hellman parameters 
`openssl dhparam -out dh2048.pem 2048`