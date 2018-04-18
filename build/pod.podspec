Pod::Spec.new do |spec|
  spec.name         = 'Gwsh'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/wiseplat/go-wiseplat'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Wiseplat Client'
  spec.source       = { :git => 'https://github.com/wiseplat/go-wiseplat.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gwsh.framework'

	spec.prepare_command = <<-CMD
    curl https://gwshstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gwsh.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
