$(document).ready(function () {
    let path = document.location.toString();
    let viewIdx = path.lastIndexOf("/view");
    if (viewIdx < 0) {
        return;
    }
    let v1p = "/v1/post";
    let v1Idx = path.lastIndexOf(v1p);
    if (v1Idx < 0) {
        return;
    }

    let userSlash = path.substring(v1Idx + v1p.length + 1, viewIdx);
    let arr = userSlash.split("/");
    console.log(JSON.stringify(arr))
    if(arr.length != 2) {
        return;
    }
    let dname = decodeURIComponent(arr[0]);
    let post = decodeURIComponent(arr[1]);

    let uri = path.substring(0, viewIdx);
    console.log("ready\n" + uri)
    $.get(uri, function(data) {
        let body = data["body"];
        if(body) {
            // Plaintext share
            if(data["blobs"][0]["type"] === "markdown") {
                let html = marked.parse(body);
                $('#output').html(html);
            }
        } else {
            // Encrypted share
        }
        console.log(body)
        console.log(data)
    })

    $("#decrypt-btn").click(function(e) {
        let pp = $('#passphrase').val();
        console.log(`Got passphrase ${pp}`);
    })
});
