FROM rabbitmq:4.0.2-management

RUN apt-get update

# Install curl for download plugins
RUN apt-get install -y curl

# Download elixir dependency for message deduplication plugin
RUN curl -L -o $RABBITMQ_HOME/plugins/elixir-1.16.3.ez \
    https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.6.4/elixir-1.16.3.ez

# Download message deduplication plugin
RUN curl -L -o $RABBITMQ_HOME/plugins/rabbitmq_message_deduplication-0.6.4.ez \
    https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.6.4/rabbitmq_message_deduplication-0.6.4.ez

# Download delayed message exchange plugin
RUN curl -L -o $RABBITMQ_HOME/plugins/rabbitmq_delayed_message_exchange-4.0.2.ez \
    https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v4.0.2/rabbitmq_delayed_message_exchange-4.0.2.ez

## Enable 3rd party plugins
RUN rabbitmq-plugins enable --offline \
    elixir  \
    rabbitmq_message_deduplication  \
    rabbitmq_delayed_message_exchange

CMD ["rabbitmq-server"]