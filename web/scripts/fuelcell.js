function setUpFuelCellGauges() {
    let controlsHeight = window.innerHeight / 4;
    let controlsWidth = window.innerWidth  / 4;
    let gaugeRadius = controlsWidth / 2;
    if (controlsWidth > controlsHeight) {
        gaugeRadius = controlsHeight / 2;
    }
    let gaugeDiameter = gaugeRadius * 2;
    $("#fcPressures").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0, 0.0, 0.0],
        min: 0,
        max: 5000,
        animationDuration: 500,
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
                        break;
                    case 1 : return ("Air = " + value + 'mbar');
                        break;
                    case 2 : return ("Coolant = " + value + 'mbar');
                        break;
                    default : return ("H2/air = " + value + 'mbar');
                };
            }}
    });
    $("#fcTemperatures").jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0, 0.0, 0.0],
        min: 15,
        max: 95,
        animationDuration: 500,
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
                        break;
                    case 1 : return ("Outlet = " + value + '&#8451;');
                        break;
                    case 2 : return ("Air = " + value + '&#8451;');
                        break;
                    default : return ("Ambient = " + value + '&#8451;');
                };
            }}
    });
    $('#fcVoltages').jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0],
        min: 0,
        max: 85,
        animationDuration: 500,
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
                        break;
                    default : return ("Out = " + value + 'V');
                };
            }}
    });
    $('#fcCurrent').jqxBarGauge({
        width: controlsWidth,
        height: controlsHeight,
        values: [0.0, 0.0],
        min: 0,
        max: 150,
        animationDuration: 500,
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
                        break;
                    default : return ("Out = " + value + 'A');
                };
            }}
    });
    $('#fcStackPower').jqxGauge({
        height: gaugeDiameter - 50,
        width: gaugeDiameter - 50,
        radius: gaugeRadius - 25,
        ticksMinor: {interval: 500, size: '5%'},
        ticksMajor: {interval: 1000,size: '9%'},
        labels: {interval:4000},
        min: 0,
        max: 12000,
        value: 0,
        radius: '50%',
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Power', position: 'bottom', offset: [0, 10], visible: true},
    });
    $('#fcStackCurrent').jqxGauge({
        height: gaugeDiameter - 50,
        width: gaugeDiameter - 50,
        radius: gaugeRadius - 25,
        ticksMinor: {interval: 5, size: '5%'},
        ticksMajor: {interval: 20,size: '9%'},
        labels: {interval:20},
        min: 0,
        max: 150,
        value: 0,
        radius: '50%',
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Current', position: 'bottom', offset: [0, 10], visible: true},
    });
    $('#fcStackVolts').jqxGauge({
        height: gaugeDiameter - 50,
        width: gaugeDiameter - 50,
        radius: gaugeRadius - 25,
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
        radius: '50%',
        animationDuration: 500,
        cap: {size: '5%', style: { fill: '#ff0000', stroke: '#00ff00' }, visible: true},
        caption: {value: 'Stack Voltage', position: 'bottom', offset: [0, 10], visible: true},
    });
}

function UpdateGauges(jsonData) {
    $("#fcPressures").val([jsonData.PanFuelCellStatus.H2Pressure,
        jsonData.PanFuelCellStatus.AirPressure,
        jsonData.PanFuelCellStatus.CoolantPressure,
        jsonData.PanFuelCellStatus.H2AirPressureDiff]);
    $("#fcTemperatures").val([jsonData.PanFuelCellStatus.CoolantInletTemp,
        jsonData.PanFuelCellStatus.CoolantOutletTemp,
        jsonData.PanFuelCellStatus.AirTemp,
        jsonData.PanFuelCellStatus.AmbientTemp]);
    $("#fcVoltages").val([jsonData.PanFuelCellStatus.DCInVolts,
        jsonData.PanFuelCellStatus.DCOutVolts]);
    $("#fcCurrent").val([jsonData.PanFuelCellStatus.DCInAmps,
        jsonData.PanFuelCellStatus.DCOutAmps]);
    $("#fcStackPower").val(jsonData.PanFuelCellStatus.StackPower);
    $("#fcStackVolts").val(jsonData.PanFuelCellStatus.StackVolts);
    $("#fcStackCurrent").val(jsonData.PanFuelCellStatus.StackCurrent);
    $("#BMSPower").text(jsonData.PanFuelCellStatus.BMSPower);
    $("#BMSHigh").text(jsonData.PanFuelCellStatus.BMSHigh);
    $("#BMSLow").text(jsonData.PanFuelCellStatus.BMSLow);
    $("#BMSCurrentPower").text(jsonData.PanFuelCellStatus.BMSCurrentPower);
    $("#BMSTargetPower").text(jsonData.PanFuelCellStatus.BMSTargetPower);
    $("#BMSTargetHigh").text(jsonData.PanFuelCellStatus.BMSTargetHigh);
    $("#BMSTargetLow").text(jsonData.PanFuelCellStatus.BMSTargetLow);
    $("#FCStatus").text(jsonData.PanFuelCellStatus.RunStatus);
    let alarmText = "";
    let alarmDiv = $("#fcAlarms");
    if (jsonData.PanFuelCellStatus.Alarms.length > 0) {
        alarmText = '<span class="alarm">';
        alarmText += jsonData.PanFuelCellStatus.Alarms.join('</span><br /><span class="alarm">')
        alarmText += '</span>'
        alarmDiv.html(alarmText);
        alarmDiv.show();
    } else {
        alarmDiv.hide();
    }
}