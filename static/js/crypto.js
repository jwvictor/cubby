
// Use this to derive the key: https://github.com/tonyg/js-scrypt
// Then aes.js

var scryptClient;
var nacl;

function initScrypt() {
    scrypt_module_factory(function (scrypt) {
        // alert(scrypt.to_hex(scrypt.random_bytes(16)));
        scryptClient = scrypt;
    });
}

function initNacl() {
    nacl_factory.instantiate(function (naclInst) {
        nacl = naclInst;
    });
}

function deriveKey(passphrase) {
    let password = scryptClient.encode_utf8(passphrase);
    let salt = scryptClient.encode_utf8("cbbc");
    // var keyBytes = scryptClient.crypto_scrypt(password, salt, 32768, 8, 1, 32);
    var keyBytes = scryptClient.crypto_scrypt(password, salt, 16384, 8, 1, 32);
    return keyBytes;
    // alert(scryptClient.to_hex(keyBytes))
}

function _base64ToArrayBuffer(base64) {
    var binary_string = window.atob(base64);
    var len = binary_string.length;
    var bytes = new Uint8Array(len);
    for (var i = 0; i < len; i++) {
        bytes[i] = binary_string.charCodeAt(i);
    }
    return bytes.buffer;
}

function subarray(arr, start, end) {
    let out = [];
    for(let i = start; i < end; i++) {
        out.push(arr[i]);
    }
    return out;
}

function aesDecryptNacl(encData, key) {
    let blockSz = 16;
    let nonce = subarray(encData, 0, blockSz);
    let ciphertext = subarray(encData, blockSz, encData.length);
    let plaintext = nacl.crypto_stream_xor(ciphertext, nonce, key);
    return plaintext;
}

function aesDecrypt(encData, key) {
    let blockSz = 16;
    let iv = subarray(encData, 0, blockSz);
    let ciphertext = subarray(encData, blockSz, encData.length);
    console.log("encData: ", encData)
    console.log("key: ", key)
    console.log("iv: ", iv)
    let aesOfb = new aesjs.ModeOfOperation.ofb(key, iv);
    let decryptedBytes = aesOfb.decrypt(ciphertext);
    return decryptedBytes;
}

function aesDecryptAESJS(encData, key) {
    // Convert text to bytes
    var text = 'Text may be any length you wish, no padding is required.';
    // var textBytes = aesjs.utils.utf8.toBytes(text);

// The counter is optional, and if omitted will begin at 1
    var aesCtr = new aesjs.ModeOfOperation.ctr(key, new aesjs.Counter(5));
    // var encryptedBytes = aesCtr.encrypt(textBytes);
    var decryptedBytes = aesCtr.decrypt(encryptedBytes);

// Convert our bytes back into text
    var decryptedText = aesjs.utils.utf8.fromBytes(decryptedBytes);
    console.log(decryptedText);
// "Text may be any length you wish, no padding is required."

}