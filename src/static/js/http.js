function handleReadyStateChange(xmlHttp, callback) {
    if (xmlHttp.readyState == 4) {
        if (xmlHttp.status == 200)
            callback(xmlHttp.responseText);
        else if (xmlHttp.status == 401)
            alert("You are not admin QQ");
    }
}

function httpGetAsync(theUrl, callback) {
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function () {
        handleReadyStateChange(xmlHttp, callback);
    }
    xmlHttp.open("GET", theUrl, true); // true for asynchronous
    xmlHttp.send(null);
}

function httpPostJsonAsync(url, jsonObject, callback) {
    var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
    xmlHttp.onreadystatechange = function () {
        handleReadyStateChange(xmlHttp, callback);
    }
    xmlhttp.open("POST", url, true);  // true for asynchronous
    xmlhttp.setRequestHeader("Content-Type", "application/json");

    var jsonString = (jsonObject == null) ? null : JSON.stringify(jsonObject);
    xmlhttp.send(jsonString);
}