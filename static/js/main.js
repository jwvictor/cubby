
var encryptedBody;
var currentBlob;

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
    $.get(uri, function(data) {
        currentBlob = data;
        let body = data["body"];
        let encBody = data["encrypted_body"];
        if(body) {
            // Plaintext share
            if(data["blobs"][0]["type"] === "markdown") {
                let html = marked.parse(body);
                $('#output').html(html);
            } else {
                $('#output').text(body);
            }
        } else if(encBody) {
            // Encrypted share
            let bytes = _base64ToArrayBuffer(encBody);
            console.log("Decoded bytes: ", bytes)
            encryptedBody = new Uint8Array(bytes);
            $('#output').html("<pre>Encrypted bytes</pre>");
        }
        console.log(data)
    })

    initScrypt();
    initNacl();

    $("#decrypt-btn").click(function(e) {
        let pp = $('#passphrase').val();
        let key = deriveKey(pp);
        let ptxt = aesDecrypt(encryptedBody, key);
        console.log("plaintext ",  ptxt);


        if(currentBlob["blobs"][0]["type"] === "markdown") {
            let decryptedText = aesjs.utils.utf8.fromBytes(ptxt);
            let html = marked.parse(decryptedText);
            $('#output').html(html);
        } else {
            $('#output').text(ptxt);
        }
        $('#secrets').hide();
    })
});
