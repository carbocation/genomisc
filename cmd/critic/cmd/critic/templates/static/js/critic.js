function submitAnnotation(entry) {
    // var base64Image = CanvasToBMP.toDataURL(document.getElementById('imgCanvas'));
    var name = entry.name;

    console.log(name);

    // document.getElementById("imgBase64").value = base64Image;
    // document.getElementById("saveImage").submit();
}

// Perform actions with the keyboard
$(document).on("keydown", function(event){
    var hit = false;
    if(event.key == "t"){
        // Toggles the overlay 
        
        var url = new URL(document.location);
        var overlay = url.searchParams.get("overlay");
        console.log(overlay);

        var newOverlay = "off";
        if(overlay == "off") {
            newOverlay = "on";
        }

        var newURL = updateURLParameter(document.location.href, "overlay", newOverlay);

        window.location = newURL;
    }

    if(hit) {
        event.preventDefault();
    }
});

/**
 * http://stackoverflow.com/a/10997390/11236
 */
function updateURLParameter(url, param, paramVal){
    var newAdditionalURL = "";
    var tempArray = url.split("?");
    var baseURL = tempArray[0];
    var additionalURL = tempArray[1];
    var temp = "";
    if (additionalURL) {
        tempArray = additionalURL.split("&");
        for (var i=0; i<tempArray.length; i++){
            if(tempArray[i].split('=')[0] != param){
                newAdditionalURL += temp + tempArray[i];
                temp = "&";
            }
        }
    }

    var rows_txt = temp + "" + param + "=" + paramVal;
    return baseURL + "?" + newAdditionalURL + rows_txt;
}