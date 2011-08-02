/*
    Copyright (c) 2011 Ross Light.
    Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

    This file is part of goray.

    goray is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    goray is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

function StatusFeed(host, jobName)
{
    this.jobName = jobName;
    this.host = host;
    this.sock = new WebSocket("ws://" + host + "/status");
    this.dataReceived = "";
    this.dataSectionStarted = false;

    this.sock.onopen = jQuery.proxy(this, "onSocketOpen");
    this.sock.onclose = jQuery.proxy(this, "onSocketClose");
    this.sock.onmessage = jQuery.proxy(this, "onSocketMessage");
}

StatusFeed.prototype.onSocketOpen = function()
{
    this.sock.send(jobName + "\r\n");
};

StatusFeed.prototype.onSocketClose = function()
{
    if (this.onClose)
        this.onClose();
};

StatusFeed.prototype.onSocketMessage = function(evt)
{
    this.dataReceived = this.dataReceived + evt.data;
    var nlIndex = this.dataReceived.indexOf("\r\n");
    while (nlIndex != -1)
    {
        var line = this.dataReceived.substr(0, nlIndex);
        this.dataReceived = this.dataReceived.substr(nlIndex + 2);
        // Try to interpret as a code
        var code = Number(line.split(" ", 1)[0]);
        if (!this.dataSectionStarted && line == "")
        {
            this.dataSectionStarted = true;
        }
        else if (!this.dataSectionStarted && isFinite(code))
        {
            if (this.onCode)
                this.onCode(code);
        }
        else
        {
            // Data line
            this.dataSectionStarted = true;
            if (this.onData)
                this.onData(line);
        }
        // Advance to next line
        nlIndex = this.dataReceived.indexOf("\r\n");
    }
};
