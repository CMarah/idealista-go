# Idealista Go info retriever

This script produces a CSV file with info about multiple listings on Idealista.
The process is:

1) Download the drive file, using Google Drive's Go API.
2) Use Idealista's API to fetch missing info (max. 100 fetches a month).
3) Creates CSV with my custom format containing the new data.
4) TODO: Upload data to same file in Drive.

Environment Variables:
- `IDEALISTA_GO_FILE_ID`
- `IDEALISTA_GO_API_USER`
- `IDEALISTA_GO_API_PASS`


<br>
To set up Google Drive's API and obtain your credentials, [follow theses steps](https://developers.google.com/drive/api/v3/quickstart/go).
