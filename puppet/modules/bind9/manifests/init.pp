class bind9 {
  package { "bind9":
    ensure => latest,
  }

  file { [ "/home/qualificati/etc", "/home/qualificati/etc/bind" ]:
    ensure  => "directory",
    require => Package["bind9"],
    owner   => "qualificati",
    group   => "bind",
    mode    => 750,
  }

  file { "/home/qualificati/etc/bind/named.conf":
    ensure  => file,
    require => [
      File["/home/qualificati/etc"],
      File["/home/qualificati/etc/bind"],
    ],
    source  => "puppet:///modules/bind9/named.conf",
    owner   => "qualificati",
    group   => "bind",
    mode    => 644,
  }

  file { "/etc/bind/named.conf":
    ensure  => link,
    require => File["/home/qualificati/etc/bind/named.conf"],
    target  => "/home/qualificati/etc/bind/named.conf",
    owner   => "root",
    group   => "bind",
  }

  file { "/var/log/named":
    ensure  => "directory",
    require => Package["bind9"],
    owner   => "root",
    group   => "bind",
    mode    => 775,
  }

  file { "/etc/apparmor.d/usr.sbin.named":
    ensure  => file,
    require => File["/etc/bind/named.conf"],
    source  => "puppet:///modules/bind9/usr.sbin.named",
    notify  => Service["apparmor"],
    owner   => "root",
    group   => "root",
    mode    => 644,
  }

  service { "apparmor":
    ensure     => running,
    enable     => true,
    hasstatus  => true,
    hasrestart => true,
  }

  service { "bind9":
    ensure     => running,
    enable     => true,
    hasstatus  => true,
    hasrestart => true,
    subscribe  => [
      File["/etc/bind/named.conf"],
      File["/var/log/named"],
      File["/etc/apparmor.d/usr.sbin.named"],
    ],
  }
}
