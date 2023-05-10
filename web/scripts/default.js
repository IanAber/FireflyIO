var wstimeout;

function showTimeoutMessage() {
    $("#connection").show();
    registerWebSocket();
    setInterval(() => {
        UpdateGauges(jsonData);
    }, 1000);
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
            $("#acvolts").text(jsonData.ACVolts);
            $("#acamps").text(jsonData.ACAmps);
            $("#acwatts").text(jsonData.ACWatts);
            $("#acwhr").text(jsonData.ACWattHours);
            $("#achz").text(jsonData.ACHertz);
            $("#acpf").text(jsonData.ACPowerFactor);
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

    setInterval(() => {
        UpdateGauges(jsonData);
    }, 1000);
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
