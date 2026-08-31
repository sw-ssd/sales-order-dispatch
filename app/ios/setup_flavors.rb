# 一次性:iOS dev/prod flavor 組態注入(D29)。
# 複製 Debug/Release/Profile 為 -dev/-prod 變體,各指向對應 xcconfig;
# dev 變體加 bundle id 後綴。冪等:已存在的組態跳過。
require 'xcodeproj'

proj_path = File.expand_path('../Runner.xcodeproj', __FILE__)
proj = Xcodeproj::Project.open(proj_path)

flutter_group = proj.main_group.find_subpath('Flutter', true)
%w[dev prod].each do |flavor|
  %w[Debug Release Profile].each do |mode|
    fname = "#{mode}-#{flavor}.xcconfig"
    flutter_group.new_file(fname) unless flutter_group.files.any? { |f| f.path == fname }
  end
end

def duplicate_configs(owner, proj)
  owner.build_configurations.dup.each do |config|
    %w[dev prod].each do |flavor|
      name = "#{config.name}-#{flavor}"
      next if owner.build_configurations.any? { |c| c.name == name }
      dup = owner.build_configuration_list.build_configurations.class
      new_config = proj.new(Xcodeproj::Project::Object::XCBuildConfiguration)
      new_config.name = name
      new_config.build_settings = config.build_settings.dup
      file_ref = proj.main_group.find_subpath('Flutter').files.find { |f| f.path == "#{name}.xcconfig" }
      new_config.base_configuration_reference = file_ref if file_ref
      owner.build_configuration_list.build_configurations << new_config
    end
  end
end

duplicate_configs(proj, proj)
proj.targets.each { |t| duplicate_configs(t, proj) }

# dev:bundle id 後綴 + 顯示名
proj.targets.each do |t|
  t.build_configurations.each do |c|
    next unless c.name.end_with?('-dev')
    c.build_settings['PRODUCT_BUNDLE_IDENTIFIER'] = '$(PRODUCT_BUNDLE_IDENTIFIER).dev' unless t.name == 'RunnerTests'
    c.build_settings['PRODUCT_NAME'] = 'Runner Dev' if t.name == 'Runner'
  end
end

proj.save
puts 'OK'
