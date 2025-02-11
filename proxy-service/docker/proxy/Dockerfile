FROM debian:latest

# Update and install necessary packages
RUN apt-get update && \
    apt-get install -y tor curl netcat-openbsd

# Add the torrc configuration file
ADD proxy-service/docker/proxy/config/torrc /etc/tor/torrc

# Add the verification script
ADD proxy-service/docker/proxy/config/check.sh /usr/local/bin/check.sh
RUN chmod +x /usr/local/bin/check.sh

EXPOSE 9050 9051

CMD ["sh", "-c", "tor & sleep 10 && /usr/local/bin/check.sh"]