var MILVUS_URI = "http://127.0.0.1:9091/api/v1"

function renderNodesMetrics(data) {
    let tableHTML = '<thead class="thead-light"><tr>' +
        '<th scope="col">Node Name</th>' +
        '<th scope="col">CPU Usage</th>' +
        '<th scope="col">Usage/Memory(GB)</th>' +
        '<th scope="col">Usage/Disk(GB)</th> '+
        '<th scope="col">IO Wait</th>' +
        '<th scope="col">RPC Ops/s</th>' +
        '<th scope="col">Network Throughput(MB/s)</th>' +
        '<th scope="col">Disk Throughput(MB/s)</th>' +
        '</tr></thead>';

    tableHTML += '<tbody>';
    data.nodes_info.forEach(node => {
        tableHTML += '<tr>';
        tableHTML += `<td>${node.infos['name']}</td>`;
        let hardwareInfo = node.infos['hardware_infos']
        let cpuUsage = parseFloat(`${hardwareInfo['cpu_core_usage']}`).toFixed(2)
        // tableHTML += `<td>${cpuUsage}%/${hardwareInfo['cpu_core_count']}</td>`;
        tableHTML += `<td>${cpuUsage}%</td>`;
        let memoryUsage = (parseInt(`${hardwareInfo['memory_usage']}`)  / 1024 / 1024 / 1024).toFixed(2)
        let memory = (parseInt(`${hardwareInfo['memory']}`)  / 1024 / 1024 / 1024).toFixed(2)
        tableHTML += `<td>${memoryUsage}/${memory}</td>`;
        let diskUsage = (parseInt(`${hardwareInfo['disk_usage']}`) / 1024 / 1024 / 1024).toFixed(2)
        let disk = (parseInt(`${hardwareInfo['disk']}`)  / 1024 / 1024 / 1024).toFixed(2)
        tableHTML += `<td>${diskUsage}/${disk}</td>`;
        tableHTML += `<td>0.00</td>`;
        tableHTML += `<td>100</td>`;
        tableHTML += `<td>5</td>`;
        tableHTML += `<td>20</td>`;
        tableHTML += '</tr>';
    });
    tableHTML += '</tbody>';
    document.getElementById('nodeMetrics').innerHTML = tableHTML;
}

function renderComponentInfo(data) {
    let tableHTML = `
        <thead class="thead-light">
            <tr>
                <th scope="col">Name</th>
                <th scope="col">IP</th>
                <th scope="col">Start Time</th>
                <th scope="col">State</th>
                <th scope="col">Reason</th>
            </tr>
        </thead><tbody>`;

    // Process each node's information
    data.nodes_info.forEach(node => {
        const hardwareInfo = node.infos['hardware_infos'];
        const tr = `
            <tr>
                <td>${node.infos['name']}</td>
                <td>${hardwareInfo['ip']}</td>
                <td>${node.infos['created_time']}</td>
                <td>Healthy</td>
                <td>None</td>
            </tr>`;
        tableHTML += tr
    });

    tableHTML += '</tbody>'
    document.getElementById("components").innerHTML = tableHTML;
}

function renderSysInfo(data) {
    let tableHTML = '<thead class="thead-light"><tr>' +
        ' <th scope="col">Attribute</th>' +
        ' <th scope="col">Value</th>' +
        ' <th scope="col">Description</th>' +
        '</tr></thead>';
    tableHTML += readSysInfo(data.nodes_info[0].infos['system_info'])

    tableHTML += '<tr>';
    tableHTML += `<td>Started Time</td>`;
    tableHTML += `<td>${data.nodes_info[0].infos['created_time']}</td>`;
    tableHTML += `<td></td></tr>`;
    tableHTML += '</tbody>';

    // Display table in the div
    document.getElementById('sysInfo').innerHTML = tableHTML;
}

function renderClientsInfo(data) {
    let tableHTML = '<thead class="thead-light"><tr>' +
        '   <th scope="col">Host</th>' +
        '   <th scope="col">User</th>' +
        '   <th scope="col">SdkType</th>' +
        '   <th scope="col">SdkVersion</th>' +
        '   <th scope="col">LocalTime</th> ' +
        '   <th scope="col">LastActiveTime</th> ' +
         '</tr></thead>';

    data.forEach(client => {
        const tr = `
            <tr>
                <td>${client['host']}</td>
                <td>${client['user'] || "default"}</td>
                <td>${client['sdk_type']}</td>
                <td>${client['sdk_version']}</td>
                <td>${client['local_time']}</td>
                <td>${client['reserved']['last_active_time']}</td>
            </tr>`;
        tableHTML += tr
    });

    tableHTML += '</tbody>'
    document.getElementById("clients").innerHTML = tableHTML;
}

function readSysInfo(systemInfo) {
    let row = ''
    row += '<tr>';
    row += `<td>GitCommit</td>`;
    row += `<td>${systemInfo.system_version}</td>`;
    row += `<td></td>`;
    row += '</tr>';

    row += '<tr>';
    row += `<td>Deploy Mode</td>`;
    row += `<td>${systemInfo.deploy_mode}</td>`;
    row += `<td></td>`;
    row += '</tr>';

    row += '<tr>';
    row += `<td>Build Version</td>`;
    row += `<td>${systemInfo.build_version}</td>`;
    row += `<td></td>`;
    row += '</tr>';

    row += '<tr>';
    row += `<td>Build Time</td>`;
    row += `<td>${systemInfo.build_time}</td>`;
    row += `<td></td>`;
    row += '</tr>';

    row += '<tr>';
    row += `<td>Go Version</td>`;
    row += `<td>${systemInfo.used_go_version}</td>`;
    row += `<td></td>`;
    row += '</tr>';
    return row
}

function renderConfigs(obj) {
    let tableHTML = '<thead class="thead-light"><tr>' +
        ' <th scope="col">Attribute</th>' +
        ' <th scope="col">Value</th>' +
        '</tr></thead>';

    Object.keys(obj).forEach(function(prop) {
        tableHTML += '<tr scope="row">';
        tableHTML += `<td>${prop}</td>`;
        tableHTML += `<td>${obj[prop]}</td>`;
        tableHTML += `</tr>`;
        tableHTML += '</tbody>';
    });
    document.getElementById('mConfig').innerHTML = tableHTML;
}