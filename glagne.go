// This is simple Dockerfile generator
package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type ParsingYaml struct {
	From     string `yaml:"FROM"`
	Composer string `yaml:"composer,omitempty"`
	PhpExt   string `yaml:"php_modules"`
}

type Version struct {
	php         string
	distrib     string
	packageName string
}

//install all-software
func SoftInstallApk() string {
	software := "RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing gnu-libiconv && \\\n"
	software += "echo @testing http://nl.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories && \\\n"
	software += "echo @main http://mirror.yandex.ru/mirrors/alpine/edge/main >>  /etc/apk/repositories && \\\n"
	software += "echo @community http://mirror.yandex.ru/mirrors/alpine/edge/community >>  /etc/apk/repositories && \\\n"
	software += "echo /etc/apk/respositories && \\\n"
	software += "apk update && \\\n"
	software += "apk add --no-cache bash \\\n"
	software += "wget \\\n"
	software += "supervisor \\\n"
	software += "curl \\\n"
	software += "libcurl \\\n"
	software += "git \\\n"
	software += "python \\\n"
	software += "python-dev \\\n"
	software += "py-pip \\\n"
	software += "augeas-dev \\\n"
	software += "openssl-dev \\\n"
	software += "ca-certificates \\\n"
	software += "dialog \\\n"
	software += "autoconf \\\n"
	software += "make \\\n"
	software += "gcc \\\n"
	software += "musl-dev \\\n"
	software += "linux-headers \\\n"
	software += "libmcrypt-dev \\\n"
	software += "libpng-dev \\\n"
	software += "icu-dev \\\n"
	software += "libpq \\\n"
	software += "libxslt-dev \\\n"
	software += "libffi-dev \\\n"
	software += "freetype-dev \\\n"
	software += "sqlite-dev \\\n"
	software += "bzip2-dev \\\n"
	software += "libmemcached-dev \\\n"
	software += "libjpeg-tubo-dev \\\n"
	software += "&& \\\n"
	return software
}
// php composer installation
func PhpComposerSetup() string {
	compose := "EXPECTED_COMPOSER_SIGNATURE=$(wget -q -O - https://composer.github.io/installer.sig) && \\\n"
	compose += "	php -r \"copy('https://getcomposer.org/installer', 'composer-setup.php');\" && \\\n"
	compose += "	php -r \"if (hash_file('SHA384', 'composer-setup.php') === '${EXPECTED_COMPOSER_SIGNATURE}') "
	compose += "{ echo 'Composer.phar Installer verified'; } else { echo 'Composer.phar Installer corrupt'; "
	compose += "unlink('composer-setup.php'); } echo PHP_EOL;\" && \\\n"
	compose += "php composer-setup.php --install-dir=/usr/bin --filename=composer && \\\n"
	compose += "php -r \"unlink('composer-setup.php');\" && \\\n"
	return compose
}
//install memcached
func StdConfAndMake() string {
	MakeAndConf := "phpize &&\\\n"
	MakeAndConf += "./configure && \\\n"
	MakeAndConf += "make && \\\n"
	MakeAndConf += "make install &&\\\n"
	return MakeAndConf
}
func InstallMemcached() string {
	memcach := "apk add --virtual .memcached-build-dependencies \\\n"
	memcach += "	libmemcached-dev \\\n"
	memcach += "	cyrus-sasl-dev && \\\n"
	memcach += "apk add --virtual .memcached-runtime-dependencies \\\n"
	memcach += "libmemcached &&\\\n"
	memcach += "git clone -o ${MEMCACHED_TAG} --depth 1 https://github.com/php-memcached-dev/php-memcached.git /tmp/php-memcached && \\\n"
	memcach += "cd /tmp/php-memcached &&\\\n"
	memcach += "phpize &&\\\n"
	memcach += "./configure \\\n"
	memcach += "    --disable-memcached-sasl \\\n"
	memcach += "    --enable-memcached-msgpack \\\n"
	memcach += "    --enable-memcached-json && \\\n"
	memcach += "make && \\\n"
	memcach += "make install && \\\n"
	memcach += "apk del .memcached-build-dependencies && \\\n"
	return memcach
}

func InstallMsgpack() string {
	msgpack := "git clone -o ${MSGPACK_TAG} --depth 1 https://github.com/msgpack/msgpack-php.git /tmp/msgpack-php && \\\n"
	msgpack += "cd /tmp/msgpack-php && \\\n"
	msgpack += StdConfAndMake()
	return msgpack
}

func InstallImagick() string {
	imagick := "apk add --no-cache --virtual .imagick-build-dependencies \\\n"
	imagick += "  autoconf \\\n"
	imagick += "  g++ \\\n"
	imagick += "  gcc \\\n"
	imagick += "  git \\\n"
	imagick += "  imagemagick-dev \\\n"
	imagick += "  libtool \\\n"
	imagick += "  make \\\n"
	imagick += "  tar && \\\n"
	imagick += "apk add --virtual .imagick-runtime-dependencies \\\n"
	imagick += "  imagemagick &&\\\n"
	imagick += "git clone -o ${IMAGICK_TAG} --depth 1 https://github.com/mkoppanen/imagick.git /tmp/imagick &&\\\n"
	imagick += "cd /tmp/imagick && \\\n"
	imagick += StdConfAndMake()
	imagick += "echo \"extension=imagick.so\" > /usr/local/etc/php/conf.d/ext-imagick.ini && \\\n"
	imagick += "apk del .imagick-build-dependencies && \\\n"
	return imagick
}
// Setup modules from code(GIT)
func UnstandartModulesInstall(module string) (string, string) {
	if module == "memcached" {
		// version memcached
		arg := "ARG MEMCACHED_TAG=v3.0.4"
		return arg, InstallMemcached()
	}
	if module == "msgpack" {
		arg := "ARG MSGPACK_TAG=msgpack-2.0.2"
		return arg, InstallMsgpack()
	}
	if module == "imagick" {
		arg := "ARG IMAGICK_TAG = \"3.4.2\""
		return arg, InstallImagick()
	}
	return "", ""
}

func main() {
	conf, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	confYaml := ParsingYaml{}

	err = yaml.Unmarshal([]byte(conf), &confYaml)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	PhpVersion := make(map[string]Version)
	PhpVersion["7.1-alpine"] = Version{
		php:         "7.1",
		distrib:     "alpine",
		packageName: "php:7.1-fpm-alpine",
	}
	PhpVersion["7.2-alpine"] = Version{
		php:         "7.1",
		distrib:     "alpine",
		packageName: "php:7.2-fpm-alpine",
	}
	PhpVersion["7.1-jessie"] = Version{
		php:         "7.1",
		distrib:     "debian",
		packageName: "php:7.1-fpm-jessie",
	}
	PhpVersion["7.2-jessie"] = Version{
		php:         "7.2",
		distrib:     "debian",
		packageName: "php:7.2-fpm-jessie",
	}

	maintainer := "\"DockerFile generator by fp <alexwolk01@gmail.com>\" \n"
	//composer := true
	var phpModules []string
	phpModules = strings.Split(confYaml.PhpExt, " ")
	ModulesNopecl := []string{"memcached", "imagick", "msgpack"}

	var switcher bool
	ARG := "\n"
	ModulesLines := ""
	var arg, StrModule string
	DockerModules := []string{"iconv", "pdo", "sqlite", "mysqli", "gd", "exif", "intl", "xsl",
		"json", "soap", "dom", "zip", "opcache", "xml", "mbstring",
		"bz2", "calendar", "ctype", "bcmatch",
	}

	DockerPhpExtInstall := "docker-php-ext-install "
	for _, module := range phpModules {
		switcher = true
		for _, nopecl := range ModulesNopecl {
			if module == nopecl {
				arg, StrModule = UnstandartModulesInstall(module)
				//generate script
				switcher = false
				ModulesLines += StrModule
				ARG += arg + "\n"
			}
		}
		if switcher {
			for _, dockerModule := range DockerModules {
				if module == dockerModule {
					DockerPhpExtInstall += module + " "
				}
			}
		}

	}

	letsencrypt := "pip install -U pip && \\\npip install -U certbot && \\\nmkdir -p /etc/letsencrypt/webrootauth && \\\n"
	DockerPhpExtInstall += "&& \\\ndocker-php-source delete && \\\n"
	ARG += "\n"
	clean := "apk del gcc musl-dev linux-headers libffi-dev augeas-dev python-dev make autoconf \n"
	ENV := "ENV php_conf /usr/local/etc/php-fpm.conf\n"
	ENV += "ENV fpm_conf /usr/local/etc/php-fpm.d/www.conf\n"
	ENV += "ENV php_vars /usr/local/etc/php/conf.d/docker-vars.ini\n"
	ENV += "ENV LD PRELOAD /usr/lib/preloadable_libconv.so php\n"
	HEAD := "FROM " + PhpVersion[confYaml.From].packageName + "\n" + "LABEL maintainer = " + maintainer + "\n"

	Dockerfile := HEAD
	Dockerfile += ENV
	Dockerfile += ARG
	if PhpVersion[confYaml.From].distrib == "alpine" {
		Dockerfile += SoftInstallApk()
	}
	Dockerfile += DockerPhpExtInstall
	if confYaml.Composer == "YES" {
		Dockerfile += PhpComposerSetup()
	}
	Dockerfile += ModulesLines
	Dockerfile += letsencrypt
	Dockerfile += clean
	fmt.Println(Dockerfile)
}
