#!/usr/bin/env ruby
# frozen_string_literal: true

require 'octokit'
require 'digest'

repository = 'automattic/hostmgr-packer-plugin'

abort 'No GitHub Token present in environment' if ENV['GITHUB_TOKEN'].nil?

## Find the tag
tag_name = ENV.fetch('BUILDKITE_TAG', nil)
abort 'No `BUILDKITE_TAG` environment variable set' if ENV['BUILDKITE_TAG'].nil?

## Find the release (creating it, if necessary)
client = Octokit::Client.new(access_token: ENV.fetch('GITHUB_TOKEN', nil))
release = client.releases(repository).select { |gh_release| tag_name == gh_release[:tag_name] }.first
release = client.create_release(repository, tag_name) if release.nil?

HASHES_FILE = "packer-plugin-hostmgr_#{tag_name}_SHA256SUMS"

[
  "packer-plugin-hostmgr_#{tag_name}_x5.0_darwin_amd64.zip",
  "packer-plugin-hostmgr_#{tag_name}_x5.0_darwin_arm64.zip"
].map do |filename|
  filepath = File.join(Dir.pwd, filename)
  abort "No file found at `#{filepath}`" unless File.file?(filepath)

  client.upload_asset(
    release[:url],
    filepath,
    content_type: 'application/octet-stream'
  )

  sha256 = Digest::SHA256.file filepath

  File.write(HASHES_FILE, "#{sha256.hexdigest}  #{filename}\n", mode: 'a')
end

client.upload_asset(
  release[:url],
  HASHES_FILE,
  content_type: 'text/plain'
)
