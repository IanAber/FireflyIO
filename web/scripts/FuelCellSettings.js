function updateDisplay(data) {
    $("#PowerDemand").val(data.BMSTargetPower.toFixed(1));
    $("#HighBattDemand").val(data.BMSTargetHigh.toFixed(1));
    $("#LowBattDemand").val(data.BMSTargetLow.toFixed(1));
//    $("#SwitchOff").hide();
}

function getSettings() {
    $.get("getFuelCell", function(data, status){
        if (status == "success") {
            updateDisplay(data);
            $("#updating").hide();
        } else {
            alert("Status = " + status);
        }
    });
 return true;
}

function PowerDown() {
    val = parseFloat($("#PowerDemand").val());
    if (val <= 0) {
        return;
    }
    $("#PowerDemand").val((val - 0.1).toFixed(1))
}

function PowerUp() {
    val = parseFloat($("#PowerDemand").val());
    if (val >= 10) {
        return;
    }
    $("#PowerDemand").val((val + 0.1).toFixed(1))
}

function HighBattUp() {
    val = parseFloat($("#HighBattDemand").val());
    if (val >= 70) {
        return;
    }
    $("#HighBattDemand").val((val + 0.1).toFixed(1));
}

function HighBattDown() {
    val = parseFloat($("#HighBattDemand").val());
    if (val <= 35) {
        return;
    }
    $("#HighBattDemand").val((val - 0.1).toFixed(1));
}

function LowBattUp() {
    val = parseFloat($("#LowBattDemand").val());
    if (val >= 70) {
        return;
    }
    $("#LowBattDemand").val((val + 0.1).toFixed(1));
}

function LowBattDown() {
    val = parseFloat($("#LowBattDemand").val());
    if (val <= 35) {
        return;
    }
    $("#LowBattDemand").val((val - 0.1).toFixed(1));
}

function StartFuelCell() {
    $.ajax({
        method : "PUT",
        url: "/setFuelCell/Start"
    });
    alert("Starting Fuel Cell");
}

function StopFuelCell() {
    $.ajax({
        method : "PUT",
        url: "/setFuelCell/Stop"
    });
    alert("Stopping fuel cell");
}

function UpdateFuelCell() {
    $("#updating").show();
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
}

function showTimeoutMessage() {
    $("#connection").show();
    registerWebSocket();
}

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
            jsonData.Relays.Relays.forEach(UpdateRelay);
            jsonData.DigitalIn.Inputs.forEach(UpdateInput);
            jsonData.DigitalOut.Outputs.forEach(UpdateOutput);
            jsonData.Analog.Inputs.forEach(UpdateAnalog);
            $("#acvolts").text(jsonData.ACVolts);
            $("#acampss").text(jsonData.ACAmps);
            $("#acwatts").text(jsonData.ACWatts);
            $("#acwhr").text(jsonData.ACWattHours);
            $("#achz").text(jsonData.ACHertz);
            $("#acpf").text(jsonData.ACPowerFactor);
            UpdateGauges(jsonData);

        } catch (e) {
            alert(e);
        }
    }
}
