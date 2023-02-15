# Downloading images from Funda in max res

Update! I made it into a tool: https://funda-image-downloader-tvorcx4jbq-ew.a.run.app/

## Old manual instructions
Go to Funda page, and hit "Alle media" ("All media").
Paste this in the console, and click back to focus the page within 5s:

```
setTimeout(() => {
  navigator.clipboard.writeText(Array.from(document.querySelectorAll("#overview-photos img"))
    .map((i) => i.src)
    .filter((v) => v.indexOf("cloud") >= 0)
    .map((v) => v.replace("_1440x960", "_2160")
  ).join("\n");
  console.log("copied to clipboard");
}), 5000)
```

Then use wget, and it will start downloading all images.
```
pbpaste | wget -i -
```
