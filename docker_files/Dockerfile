# Use alpine with golang as base image
FROM alpine:latest

# Set root as working dir.
WORKDIR /

# Create dir. for go executable
RUN mkdir -p /opt/cnis

# Copy go executable into container
ADD go_bin/cnis /opt/cnis

# Run go executable
CMD /opt/cnis/cnis
