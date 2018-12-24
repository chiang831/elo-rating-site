function handleReadyStateChange(xmlHttp, succeessCallback) {
    if (xmlHttp.readyState == 4) {
        if (xmlHttp.status == 200)
            succeessCallback(xmlHttp.responseText);
        else if (xmlHttp.status == 401)
            alert("You are not admin QQ");
        else if (xmlHttp == 500) {
            alert(xmlHttp.status + "Internal Failure:\n" + xmlHttp.responseText)
        }
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
    var xmlHttp = new XMLHttpRequest();   // new HttpRequest instance 
    xmlHttp.onreadystatechange = function () {
        handleReadyStateChange(xmlHttp, callback);
    }
    xmlHttp.open("POST", url, true);  // true for asynchronous
    xmlHttp.setRequestHeader("Content-Type", "application/json");

    var jsonString = (jsonObject == null) ? null : JSON.stringify(jsonObject);
    xmlHttp.send(jsonString);
}