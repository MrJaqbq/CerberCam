﻿using System.IO;
using System.Threading.Tasks;
using ProtoBuf;
using RabbitMQ.Client;

namespace CerberCam.Core
{
    public class QueueManager : IQueueWrapper
    {
        private const string hostName = "cerbercam.cloudapp.net";
        private const string queueName = "picturesQueue";

        private readonly IConnectionFactory _factory;

        public QueueManager() : this(new ConnectionFactory { HostName = hostName })
        {
        }

        public QueueManager(IConnectionFactory connectionFactory)
        {
            _factory = connectionFactory;
        }

        public void SendAsync(ref Message msg)
        {
            using (IConnection connection = _factory.CreateConnection())
            {
                using (IModel channel = connection.CreateModel())
                {
                    channel.QueueDeclare(queue: queueName,
                                 durable: false,
                                 exclusive: false,
                                 autoDelete: false,
                                 arguments: null);

                    byte[] data = SerializeMessage(ref msg);

                    channel.BasicPublish(exchange: "",
                                 routingKey: queueName,
                                 basicProperties: null,
                                 body: data);
                }
            }
        }

        private byte[] SerializeMessage(ref Message msg)
        {
            using (MemoryStream stream = new MemoryStream())
            {
                Serializer.Serialize(stream, msg);
                return stream.ToArray();
            }
        }
    }
}
