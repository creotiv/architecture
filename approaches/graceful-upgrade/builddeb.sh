go build -o graceful-upgrade-example server.go

systemd_service="systemd/graceful-upgrade-example@.service"

fpm_args=()
# Adding systemd service task
fpm_args+=(--deb-systemd "$systemd_service")
# Adding systemd pre and post install scripts
# they will be start and stop our application
fpm_args+=(--before-upgrade "$systemd_service.preinst")
fpm_args+=(--before-install "$systemd_service.preinst")
fpm_args+=(--after-upgrade "$systemd_service.postinst")
fpm_args+=(--after-install "$systemd_service.postinst")
# All start/stop in pre|postinst files
fpm_args+=(--no-deb-systemd-auto-start)
fpm_args+=(--no-deb-systemd-restart-after-upgrade)

timestamp=$(date +%s)

mkdir -p build | true

fpm "${fpm_args[@]}" \
    --verbose \
    --force \
    --input-type dir \
    --output-type deb \
    --package "build/graceful-upgrade-example.deb" \
    --prefix /opt/graceful-upgrade-example/bin \
    --name "graceful-upgrade-example" \
    --version "${timestamp}" \
    --architecture native \
    graceful-upgrade-example
