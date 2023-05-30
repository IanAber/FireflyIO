var wstimeout;

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
            jsonData.Relays.Relays.forEach(UpdateRelay);
            jsonData.DigitalIn.Inputs.forEach(UpdateInput);
            jsonData.DigitalOut.Outputs.forEach(UpdateOutput);
            jsonData.Analog.Inputs.forEach(UpdateAnalog);
            if (jsonData.ACMeasurements.length > 0) {
                $("#ACMeasurementsDiv").show();
                jsonData.ACMeasurements.forEach(updateAC);
                for (i = jsonData.ACMeasurements.length + 1; i < 5; i++) {
                    $("#AC" + i).hide();
                    $("#ACErr" + i).hide();
                }
            } else {
                $("#ACMeasurementsDiv").hide();
            }
            if (jsonData.DCMeasurements.length > 0) {
                $("#DCMeasurementsDiv").show();
                jsonData.DCMeasurements.forEach(updateDC);
                for (i = jsonData.DCMeasurements.length + 1; i < 5; i++) {
                    $("#DC" + i).hide();
                    $("#DCErr" + i).hide();
                }
            } else {
                $("#DCMeasurementsDiv").hide();
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
            UpdateGauges(jsonData);
        } catch (e) {
            alert(e);
        }
    }
}

function setupPage() {
    setUpFuelCellGauges();
    registerWebSocket();

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


function UpdateRelay(Relay, idx) {
    td_relay = $("#relay" + idx);
    if (Relay.On) {
        td_relay.removeClass("RelayOff");
        td_relay.removeClass("RelayChanging");
        td_relay.addClass("RelayOn");
    } else {
        td_relay.removeClass("RelayOn");
        td_relay.removeClass("RelayChanging");
        td_relay.addClass("RelayOff");
    }
    $("#relayText"+idx).text(Relay.Name);
}

function UpdateInput(Input, idx) {
    td_input = $("#di" + idx);
    if (Input.Pin) {
        td_input.removeClass("DILow");
        td_input.addClass("DIHigh");
    } else {
        td_input.removeClass("DIHigh");
        td_input.addClass("DILow");
    }
    $("#InputText"+idx).text(Input.Name);
}

function UpdateOutput(Output, idx) {
    td_output = $("#do" + idx);
    if (Output.Pin) {
        td_output.removeClass("DOLow");
        td_output.addClass("DOHigh");
    } else {
        td_output.removeClass("DOHigh");
        td_output.addClass("DOLow");
    }
    $("#OutputText"+idx).text(Output.Name);
}

function UpdateAnalog(analog, idx) {
    $("#a"+idx+"name").text(analog.Name);
    $("#a"+idx+"raw").text(analog.Raw);
    $("#a"+idx+"value").text(analog.Value.toFixed(2));
}

function updateAC(ac, idx) {
    idx++;
    if (ac.Error === "") {
        $("#AC" + idx).show();
        $("#ACErr" + idx).hide();
        $("#acname" + idx).text(ac.Name);
        $("#acvolts" + idx).text(ac.ACVolts);
        $("#acamps" + idx).text(ac.ACAmps);
        $("#acwatts" + idx).text(ac.ACWatts);
        $("#acwhr" + idx).text(ac.ACWattHours);
        $("#achz" + idx).text(ac.ACHertz);
        $("#acpf" + idx).text(ac.ACPowerFactor);
    } else {
        $("#AC" + idx).hide();
        $("#ACErr" + idx).show();
        $("#acnameerr" + idx).text(ac.Name)
        $("#acerr" + idx).text(ac.Error);
    }
}

function updateDC(dc, idx) {
    idx++;
    if (dc.Error === "") {
        $("#DC" + idx).show();
        $("#DCErr" + idx).hide();
        $("#dcname" + idx).text(dc.Name);
        $("#dcvolts" + idx).text(dc.DCVolts);
        $("#dcamps" + idx).text(dc.DCAmps);
    } else {
        $("#DC" + idx).hide();
        $("#DCErr" + idx).show();
        $("#dcnameerr" + idx).text(dc.Name)
        $("#dcerr" + idx).text(dc.Error);
    }
}


function  clickRelay(id) {
    rl = $("#relay" + id);
    if (rl.hasClass("RelayOff")) {
        action = "on";
    } else if (rl.hasClass("RelayOn")){
        action = "off";
    }
    rl.removeClass("RelayOn");
    rl.removeClass("RelayOff");
    rl.addClass("RelayChanging");
    putString = "/setRelay/"+id+"/" + action
    $.ajax({
        url: putString,
        type: 'put',
        headers: {
            "Content-Type": "application/json"
        },
        dataType: 'json'
    })
}

function  clickOutput(id) {
    op = $("#do" + id);
    if (op.hasClass("DOLow")) {
        action = "on";
    } else {
        action = "off";
    }
    putString = "/setOutput/"+id+"/" + action
    $.ajax({
        url: putString,
        type: 'put',
        headers: {
            "Content-Type": "application/json"
        },
        dataType: 'json'
    })
}
