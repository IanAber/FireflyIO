function loadSettings() {
    fetch("/getSettings")
        .then( function(response) {
            if (response.status === 200) {
                response.json()
                    .then(function (data) {
                        $("#name").val(data.Name);
                        data.AnalogChannels.forEach(SetAnalogSettings);
                        data.DigitalInputs.forEach(SetDigitalInputSettings);
                        data.DigitalOutputs.forEach(SetDigitalOuputSettings);
                        data.Relays.forEach(SetRelaySettings);
                        data.ACMeasurement.forEach(SetACMeasurementSettings);
                        data.DCMeasurement.forEach(SetDCMeasurementSettings);
                        if (data.FuelCellSettings.IgnoreIsoLow) {
                            $("#isoLowBehaviour").val("true")
                        } else {
                            $("#isoLowBehaviour").val("false")
                        }
                    });
            }
        })
}

function SetAnalogSettings(channel){
    $("#a"+channel.Port+"name").val(channel.Name);
    $("#a"+channel.Port+"LowVal").val(channel.LowerCalibrationActual);
    $("#a"+channel.Port+"LowA2D").val(channel.LowerCalibrationAtoD);
    $("#a"+channel.Port+"HighVal").val(channel.UpperCalibrationActual);
    $("#a"+channel.Port+"HighA2D").val(channel.UpperCalibrationAtoD);
}

function SetRelaySettings(channel) {
    $("#relay"+channel.Port+"name").val(channel.Name);
}

function SetDigitalOuputSettings(channel) {
    $("#do"+channel.Port+"name").val(channel.Name);
}

function SetDigitalInputSettings(channel) {
    $("#di"+channel.Port+"name").val(channel.Name);
}

function SetACMeasurementSettings(channel) {
    $("#ACMeasurement"+channel.SlaveID).val(channel.Name);
}

function SetDCMeasurementSettings(channel) {
    $("#DCMeasurement"+channel.SlaveID).val(channel.Name);
}