var wstimeout;

function updateDisplay(data) {
    $("#PowerDemand").val(data.BMSTargetPower.toFixed(1));
    $("#HighBattDemand").val(data.BMSTargetHigh.toFixed(1));
    $("#LowBattDemand").val(data.BMSTargetLow.toFixed(1));
}

function getSettings() {
    $.get("getFuelCell", function(data, status){
        if (status === "success") {
            updateDisplay(data);
        } else {
            alert("Status = " + status);
        }
    });
 return true;
}

function PowerDown() {
    let pd = $("#PowerDemand");
    val = parseFloat(pd.val());
    if (val <= 0) {
        return;
    }
    pd.val((val - 0.1).toFixed(1))
}

function PowerUp() {
    let pd = $("#PowerDemand");
    val = parseFloat(pd.val());
    if (val >= 10) {
        return;
    }
    pd.val((val + 0.1).toFixed(1))
}

function HighBattUp() {
    let hb = $("#HighBattDemand");
    val = parseFloat(hb.val());
    if (val >= 70) {
        return;
    }
    hb.val((val + 0.1).toFixed(1));
}

function HighBattDown() {
    let hb = $("#HighBattDemand");
    val = parseFloat(hb.val());
    if (val <= 35) {
        return;
    }
    hb.val((val - 0.1).toFixed(1));
}

function LowBattUp() {
    let ld = $("#LowBattDemand");
    val = parseFloat(ld.val());
    if (val >= 70) {
        return;
    }
    ld.val((val + 0.1).toFixed(1));
}

function LowBattDown() {
    let ld = $("#LowBattDemand");
    val = parseFloat(ld.val());
    if (val <= 35) {
        return;
    }
    ld.val((val - 0.1).toFixed(1));
}

function RunFuelCellClick() {
    if ($("#Enable").hasClass("swOff")) {
        alert("Control is disabled. Click the Enable button to allow control of the fuel cell.");
        return;
    }
    let btn = $("#SwitchOnOff");
    let onOff = btn.hasClass("swOn");
    btn.addClass("depressed");
    if (onOff) {
        url = "/setFuelCell/Stop";
    } else {
        url = "/setFuelCell/Start";
    }
    $.ajax({
        method : "PUT",
        url: url
    });
}

function ExhaustClick() {
    if ($("#Enable").hasClass("swOff")) {
        alert("Control is disabled. Click the Enable button to allow control of the fuel cell.");
        return;
    }
    let btn = $("#Exhaust");
    btn.addClass("depressed");
    if (btn.hasClass('swOn')) {
        url = "/setFuelCell/ExhaustClose";
    } else {
        url = "/setFuelCell/ExhaustOpen";
    }
    btn.addClass("depressed");
    $.ajax({
        method : "PUT",
        url: url
    });
}

function EnableFuelCellClick() {
    let btn = $("#Enable");
    btn.addClass("depressed");
    if (btn.hasClass("swOn")) {
        url = "/setFuelCell/Disable";
    } else {
        url = "/setFuelCell/Enable";
    }
    $.ajax({
        method : "PUT",
        url: url
    });
}

function UpdateFuelCell() {
    $("#settingsForm").submit();
}

function setupPage() {
    setUpFuelCellGauges();
    registerWebSocket();
    getSettings();

    $(window).resize(function () {
        if (this.resizeTO) clearTimeout(this.resizeTO);
        this.resizeTO = setTimeout(function () {
            $(this).trigger('windowResize');
        }, 500);
    });

    $(window).on('windowResize', function () {
        window.location.reload();
    });
    setInterval(() => {
        UpdateGauges(jsonData);
    }, 1000);
}

function showTimeoutMessage() {
    $("#connection").show();
    registerWebSocket();
}

var jsonData;

function registerWebSocket() {
    let url = "ws://" + window.location.host + "/ws";
    let conn = new WebSocket(url);
    wsTimeou = 0;

    conn.onclose = function () {
        $("#connection").show();
    }
    conn.onmessage = function (evt) {
        if (wstimeout !== 0) {
            clearTimeout(wstimeout);
            $("#connection").hide();
        }
        wstimeout = setTimeout(showTimeoutMessage, 15000)
        try {
            jsonData = JSON.parse(evt.data);

            $("#system").text(jsonData.System);
            $("#version").text(jsonData.Version);
            let sw = $("#Exhaust");
            sw.removeClass("depressed");
            bOn = (sw.attr('state') === "true");
            if (jsonData.PanFuelCellStatus.ExhaustOpen) {
                sw.addClass("swOn");
                sw.removeClass("swOff");
            } else {
                sw.addClass("swOff");
                sw.removeClass("swOn");
            }

            let onOff = $("#SwitchOnOff");
            onOff.removeClass("depressed");
            if (jsonData.PanFuelCellStatus.Start) {
                onOff.addClass("swOn");
                onOff.removeClass("swOff");
            } else {
                onOff.addClass("swOff");
                onOff.removeClass("swOn");
            }

            let en = $("#Enable");
            en.removeClass("depressed");
            if (jsonData.PanFuelCellStatus.Enable) {
                en.addClass("swOn");
                en.removeClass("swOff");
            } else {
                en.addClass("swOff");
                en.removeClass("swOn")
            }

            $("#BMSPower").text(jsonData.PanFuelCellStatus.BMSPower);
            $("#BMSHigh").text(jsonData.PanFuelCellStatus.BMSHigh);
            $("#BMSLow").text(jsonData.PanFuelCellStatus.BMSLow);
            $("#BMSCurrentPower").text(jsonData.PanFuelCellStatus.BMSCurrentPower);
            $("#BMSTargetPower").text(jsonData.PanFuelCellStatus.BMSTargetPower);
            $("#BMSTargetHigh").text(jsonData.PanFuelCellStatus.BMSTargetHigh);
            $("#BMSTargetLow").text(jsonData.PanFuelCellStatus.BMSTargetLow);
            $("#FCStatus").text(jsonData.PanFuelCellStatus.RunStatus);
            $("#FCDCOutputStatus").text(jsonData.PanFuelCellStatus.DCOutputStatus);

        } catch (e) {
            alert(e);
        }
    }
}
