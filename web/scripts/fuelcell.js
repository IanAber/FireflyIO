function smallest(x, y) {
    if ((x > y) && (y > 0)) {
        return y;
    } else {
        return x;
    }
}

function setUpFuelCellGauges() {
    // let controlsHeight = window.innerHeight / 4;
    // let controlsWidth = window.innerWidth  / 5;
    // if (controlsWidth < controlsHeight) {
    //     controlsHeight = controlsWidth;
    // } else {
    //     controlsWidth = controlsHeight;
    // }
    // let gaugeRadius = controlsWidth / 2;
    // let gaugeDiameter = gaugeRadius * 2;
    let pressures = $("#fcPressures");
    let size = smallest(pressures.width(), pressures.height());
    pressures.jqxBarGauge({
        width: size,
        height: size,
        values: [0.0, 0.0, 0.0, 0.0],
        min: 0,
        max: 150,
        animationDuration: 0,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Pressures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#000066', '#006600', '#660000', '#006666'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("H2 = " + value + 'mbar');
                    case 1 : return ("Air = " + value + 'mbar');
                    case 2 : return ("Coolant = " + value + 'mbar');
                    default : return ("H2/air = " + value + 'mbar');
                }
            }}
    });
    let temperatures= $("#fcTemperatures");
    size = smallest(temperatures.width(), temperatures.height());
    temperatures.jqxBarGauge({
        width: size,
        height: size,
        values: [0.0, 0.0, 0.0, 0.0],
        min: 5,
        max: 90,
        animationDuration: 0,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Temperatures',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#000066', '#006600', '#660000', '#006666'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("Inlet = " + value + '&#8451;');
                    case 1 : return ("Outlet = " + value + '&#8451;');
                    case 2 : return ("Air = " + value + '&#8451;');
                    default : return ("Ambient = " + value + '&#8451;');
                }
            }}
    });
    let voltages = $('#fcVoltages');
    size = smallest(voltages.width(), voltages.height());
    voltages.jqxBarGauge({
        width: size,
        height: size,
        values: [0.0, 0.0],
        min: 0,
        max: 85,
        animationDuration: 0,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Voltages',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#000066', '#006600'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("In = " + value + 'V');
                    default : return ("Out = " + value + 'V');
                }
            }}
    });
    let currents = $('#fcCurrent');
    size = smallest(currents.width(), currents.height());
    currents.jqxBarGauge({
        width: size,
        height: size,
        values: [0.0, 0.0],
        min: 0,
        max: 400,
        animationDuration: 0,
        startAngle: 265,
        endAngle: 275,
        title:	{
            text: 'Current',
            font: { size: 12, color: 'black', weight: 'bold', family:"Segoi-UI"},
            margin: { top: 0, bottom: 2, left: 0, right: 0},
            verticalAlignment: 'bottom'
        },
        labels: {
            font: {size: 8,},
            precision: 1,
        },
        colorScheme: 'customColors',
        customColorScheme: { name: 'customColors', colors: ['#000066', '#006600'] },
        tooltip: {visible: true,
            precision: 1,
            formatFunction: function (value, index){
                switch(index)
                {
                    case 0 : return("In = " + value + 'A');
                    default : return ("Out = " + value + 'A');
                }
            }}
    });
    insulation = $('#fcInsulation');
    size = smallest(insulation.width(), insulation.height());
    insulation.jqxGauge({
        height: size,
        width: size,
        radius: (size / 2) - 25,
        ticksMinor: {interval: 50, size: '5%'},
        ticksMajor: {interval: 250,size: '9%'},
        labels: {interval:250},
        min: 0,
        max: 1000,
        value: 0,
        ranges: [
            {startValue: 0, endValue: 100, style: {fill: 'RED', stroke: 'RED'}, startWidth: 9, endWidth: 7},
            {startValue: 100, endValue: 500, style: {fill: 'ORANGE', stroke: 'ORANGE'}, startWidth: 7, endWidth: 5},
            {startValue: 500, endValue: 5000, style: {fill: 'GREEN', stroke: 'GREEN'}, startWidth: 5, endWidth: 2}
        ],
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Insulation', position: 'bottom', offset: [0, 10], visible: true},
    });
    stackPower = $('#fcStackPower');
    size = smallest(stackPower.width(), stackPower.height());
    stackPower.jqxGauge({
        height: size,
        width: size,
        radius: (size / 2) - 25,
        ticksMinor: {interval: 500, size: '5%'},
        ticksMajor: {interval: 1000,size: '9%'},
        labels: {interval:5000},
        min: 0,
        max: 15000,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Power', position: 'bottom', offset: [0, 10], visible: true},
    });
    stackCurrent = $('#fcStackCurrent');
    size = smallest(stackCurrent.width(), stackCurrent.height());
    stackCurrent.jqxGauge({
        height: size,
        width: size,
        radius: (size/ 2) - 25,
        ticksMinor: {interval: 5, size: '5%'},
        ticksMajor: {interval: 50,size: '9%'},
        labels: {interval:50},
        min: 0,
        max: 400,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Current', position: 'bottom', offset: [0, 10], visible: true},
    });
    stackVolts = $('#fcStackVolts');
    size = smallest(stackVolts.width(), stackVolts.height());
    stackVolts.jqxGauge({
        height: size,
        width: size,
        radius: (size / 2) - 25,
        ranges: [
            {startValue: 0, endValue: 30, style: {fill: 'RED', stroke: 'RED'}, startWidth: 9, endWidth: 5},
            {startValue: 30, endValue: 65, style: {fill: 'GREEN', stroke: 'GREEN'}, startWidth: 5, endWidth: 5},
            {startValue: 65, endValue: 80, style: {fill: 'RED', stroke: 'RED'}, startWidth: 5, endWidth: 9}
        ],
        ticksMinor: {interval: 5, size: '5%'},
        ticksMajor: {interval: 20,size: '9%'},
        labels: {interval:10},
        min: 0,
        max: 80,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Voltage', position: 'bottom', offset: [0, 10], visible: true},
    });
    pumpSpeed = $('#fcWaterPumpSpeed');
    size = smallest(pumpSpeed.width(), pumpSpeed.height());
    pumpSpeed.jqxGauge({
        height: size,
        width: size,
        radius: (size / 2) - 25,
        ticksMinor: {interval: 500, size: '5%'},
        ticksMajor: {interval: 1000,size: '9%'},
        labels: {interval:1000},
        min: 0,
        max: 6000,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Water Pump Speed', position: 'bottom', offset: [0, 10], visible: true},
    });
    fanSpeed = $('#fcCoolingFanSpeed');
    size = smallest(fanSpeed.width(), fanSpeed.height());
    fanSpeed.jqxGauge({
        height: size,
        width: size,
        radius: (size / 2) - 25,
        ticksMinor: {interval: 500, size: '5%'},
        ticksMajor: {interval: 1000,size: '9%'},
        labels: {interval:1000},
        min: 0,
        max: 6000,
        value: 0,
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Cooling Fan Speed', position: 'bottom', offset: [0, 10], visible: true},
    });
}

function UpdateGauges(jsonData) {
    if (jsonData.PanFuelCellStatus === null) {
        return;
    }
    PressuresNewVals = [jsonData.PanFuelCellStatus.H2Pressure,
        jsonData.PanFuelCellStatus.AirPressure,
        jsonData.PanFuelCellStatus.CoolantPressure,
        jsonData.PanFuelCellStatus.H2AirPressureDiff];
    Pressures = $("#fcPressures");
    PressuresVals = Pressures.val();
    if ((PressuresVals[0] !== PressuresNewVals[0]) || (PressuresVals[1] !== PressuresNewVals[1]) ||
        (PressuresVals[2] !== PressuresNewVals[2]) || (PressuresVals[3] !== PressuresNewVals[3]))
    {
        Pressures.val(PressuresNewVals);
    }

    TemperaturesNewVals = [jsonData.PanFuelCellStatus.CoolantInletTemp,
        jsonData.PanFuelCellStatus.CoolantOutletTemp,
        jsonData.PanFuelCellStatus.AirTemp,
        jsonData.PanFuelCellStatus.AmbientTemp];
    Temperatures = $("#fcTemperatures");
    TemperaturesVals = Temperatures.val();
    if ((TemperaturesVals[0] !== TemperaturesNewVals[0]) || (TemperaturesVals[1] !== TemperaturesNewVals[1]) ||
        (TemperaturesVals[2] !== TemperaturesNewVals[2]) || (TemperaturesVals[3] !== TemperaturesNewVals[3]))
    {
        Temperatures.val(TemperaturesNewVals);
    }

    VoltagesNewVals = [jsonData.PanFuelCellStatus.DCInVolts,
        jsonData.PanFuelCellStatus.DCOutVolts];
    Voltages = $("#fcVoltages");
    VoltagesVals = Voltages.val();
    if ((VoltagesVals[0] !== VoltagesNewVals[0]) || (VoltagesVals[1] !== VoltagesNewVals[1])) {
        Voltages.val(VoltagesNewVals);
    }

    CurrentsNewVals = [jsonData.PanFuelCellStatus.DCInAmps, jsonData.PanFuelCellStatus.DCOutAmps];
    Currents = $("#fcCurrent");
    CurrentsVals = Currents.val();
    if ((CurrentsVals[0] !== CurrentsNewVals[0]) || (CurrentsVals[1] !== CurrentsNewVals[1]))
    Currents.val(CurrentsNewVals);

    if (jsonData.PanFuelCellStatus.InsulationResistance === 65535) {
        $("#InsulationDiv").hide();
    } else {
        $('#fcInsulation').val(jsonData.PanFuelCellStatus.InsulationResistance);
        $('#fcInsulationStatus').text(jsonData.PanFuelCellStatus.InsulationStatus);
        $("#fcInsulationFault").text(jsonData.PanFuelCellStatus.InsulationFault);
    }
    $("#fcStackPower").val(jsonData.PanFuelCellStatus.StackPower);
    $("#fcStackVolts").val(jsonData.PanFuelCellStatus.StackVolts);
    $("#fcStackCurrent").val(jsonData.PanFuelCellStatus.StackCurrent);
    $("#fcWaterPumpSpeed").val(jsonData.PanFuelCellStatus.WaterPumpSpeed);
    $("#fcCoolingFanSpeed").val(jsonData.PanFuelCellStatus.CoolingFanSpeed);
    var alarmText;
    let alarmDiv = $("#fcAlarms");
    if (jsonData.PanFuelCellStatus.Alarms.length > 0) {
        alarmText = '<span class="alarm">';
        alarmText += jsonData.PanFuelCellStatus.Alarms.join('</span><br /><span class="alarm">')
        alarmText += '</span>'
    } else {
        alarmText = "";
    }
    if (jsonData.PanFuelCellStatus.DCOutputStatus === "fault") {
        alarmText += '<span class="alarm">' + jsonData.PanFuelCellStatus.DCOutputFault + '</span><br />'
    }
    if (alarmText !== "") {
        alarmDiv.html(alarmText);
        alarmDiv.show();
    } else {
        alarmDiv.hide();
    }
}