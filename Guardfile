guard :go, :server => 'web.go', :timeout => 5 do
  watch(%r{\.go$})
end

guard :go, :server => 'worker.go', :timeout => 5 do
  watch(%r{\.go$})
end

notification :growl
