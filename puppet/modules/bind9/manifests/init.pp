class bind9 {
  package { "bind9":
    ensure => latest,
  }

  file { [ "/home/qualificati/etc", "/home/qualificati/etc/bind" ]:
    ensure  => "directory",
    require => Package["bind9"],
    owner   => "qualificati",
    group   => "qualificati",
    mode    => 750,
  }

  file { "/home/qualificati/etc/bind/named.conf":
    ensure  => file,
    require => [
      File["/home/qualificati/etc"],
      File["/home/qualificati/etc/bind"],
    ],
    source  => "puppet:///modules/bind9/named.conf",
  }

  file { "/etc/bind/named.conf":
    ensure  => link,
    require => File["/home/qualificati/etc/bind/named.conf"],
    target  => "/home/qualificati/etc/bind/named.conf",
  }

  service { "bind9":
    ensure     => running,
    enable     => true,
    hasstatus  => true,
    hasrestart => true,
    subscribe  => [ File["/etc/bind/named.conf"] ],
  }
}
