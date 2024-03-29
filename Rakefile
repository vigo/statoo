# frozen_string_literal: true

require 'open3'

# constants
# -----------------------------------------------------------------------------
AVAILABLE_REVISIONS = %w[major minor patch].freeze
# -----------------------------------------------------------------------------


# -----------------------------------------------------------------------------
# hidden tasks
# -----------------------------------------------------------------------------
task :command_exists, [:command] do |_, args|
  abort "#{args.command} doesn't exists" if `command -v #{args.command} > /dev/null 2>&1 && echo $?`.chomp.empty?
end

task :repo_clean do
  abort 'please commit your changes first!' unless `git status -s | wc -l`.strip.to_i.zero?
end

task :current_version do
  version_file = File.open('.bumpversion.cfg', 'r')
  data = version_file.read
  version_file.close
  match = /current_version = (\d+).(\d+).(\d+)/.match(data)
  "#{match[1]}.#{match[2]}.#{match[3]}"
end

task :has_bumpversion do
  Rake::Task['command_exists'].invoke('bumpversion')
end

task :has_gsed do
  Rake::Task['command_exists'].invoke('gsed')
end

task :bump, [:revision] => [:has_bumpversion] do |_, args|
  args.with_defaults(revision: 'patch')
  unless AVAILABLE_REVISIONS.include?(args.revision)
    abort "Please provide valid revision: #{AVAILABLE_REVISIONS.join(',')}"
  end

  system "bumpversion #{args.revision}"
end

task :get_current_branch do
  `git rev-parse --abbrev-ref HEAD`.strip
end
# -----------------------------------------------------------------------------


# default task
# -----------------------------------------------------------------------------
desc 'show avaliable tasks (default task)'
task :default do
  system('rake -sT')
end
# -----------------------------------------------------------------------------


# run tasks
# -----------------------------------------------------------------------------
namespace :test do
  desc 'run tests, generate coverage'
  task :run, [:verbose] do |_, args|
    args.with_defaults(verbose: '')
    system "go test -failfast -count=1 #{args.verbose} -coverprofile=coverage.out ./..."
  end
  
  desc "show coverage after running tests"
  task :show_coverage do
    Rake::Task["test:run"].invoke('-v')
    system "go tool cover -html=coverage.out"
  end
  
  desc "update coverage value in README"
  task :update_coverage => [:has_gsed] do
    coverage_value = `go test -count=1 -coverprofile=coverage.out ./... | grep 'ok'`.chomp.split("\t")
    coverage_ratio = coverage_value.last.split[1].gsub!('%', '%25')
    system %{
      gsed -i -r 's/coverage-[0-9\.\%]+/coverage-#{coverage_ratio}/' README.md &&
      echo "new coverage is set to: #{coverage_ratio}"
    }
  end
end

# -----------------------------------------------------------------------------


# release new version
# -----------------------------------------------------------------------------
desc "release new version #{AVAILABLE_REVISIONS.join(',')}, default: patch"
task :release, [:revision] => [:repo_clean] do |_, args|
  args.with_defaults(revision: 'patch')
  Rake::Task['bump'].invoke(args.revision)
end
# -----------------------------------------------------------------------------


# docker
# -----------------------------------------------------------------------------
DOCKER_IMAGE_TAG = 'statoo:latest'
namespace :docker do
  desc "lint Dockerfile"
  task :lint do
    system "hadolint Dockerfile"
  end
  
  desc "Run image (locally)"
  task :run, [:param] do |_, args|
    cmd_args = [args.param] + args.extras
    system %{
      docker run #{DOCKER_IMAGE_TAG} #{cmd_args.join(' ')}
    }
    exit $?.exitstatus
  end

  desc "Build image (locally)"
  task :build do
    git_commit_hash = `git rev-parse HEAD`.chomp
    goos =`go env GOOS`.chomp
    goarch =`go env GOARCH`.chomp
    build_commit_hash = "#{git_commit_hash}-#{goos}-#{goarch}"
    
    system %{
      docker build --build-arg="BUILD_INFORMATION=#{build_commit_hash}" \
        -t #{DOCKER_IMAGE_TAG} .
    }
    exit $?.exitstatus
  end
end
# -----------------------------------------------------------------------------
