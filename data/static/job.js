/*
 * job.js
 * goray
 *
 * Created by Ross Light on 2011-02-06.
 */

function StatusFeed(host, jobName)
{
    this.jobName = jobName;
    this.host = host;
    this.sock = new WebSocket("ws://" + host + "/status");
    this.dataReceived = "";
    this.code = -1;

    this.sock.onopen = jQuery.proxy(this, "onSocketOpen");
    this.sock.onmessage = jQuery.proxy(this, "onSocketMessage");
}

StatusFeed.prototype.onSocketOpen = function()
{
    this.sock.send(jobName + "\r\n");
};

StatusFeed.prototype.onSocketMessage = function(evt)
{
    this.receivedData = this.receivedData + evt.data;
    var nlIndex = this.receivedData.indexOf("\r\n");
    while (nlIndex != -1)
    {
        var line = this.receivedData.substr(0, nlIndex);
        this.receivedData = this.receivedData.substr(nlIndex + 2);
        if (this.code < 0)
        {
            this.code = Number(line.split(" ", 1)[0]);
            if (this.onReceiveCode)
            {
                this.onReceiveCode(this.code);
            }
        }
        else
        {
            if (this.onUpdate)
            {
                this.onUpdate(line);
            }
        }
        nlIndex = this.receivedData.indexOf("\r\n");
    }
};
