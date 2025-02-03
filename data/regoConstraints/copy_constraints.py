import os
def generate_constraint_files(n):
    """
    Generates n copies of each constraint file in the 'motivatingConstraints' folder
    and places them in a new 'generated_constraints' folder.
    """
    # Create the 'generated_constraints' folder if it doesn't exist
    if not os.path.exists('motivatingConstraintsCopy_'+str(n)):
        os.makedirs('motivatingConstraintsCopy_'+str(n))

    # Loop through each file in the 'motivatingConstraints' folder
    for filename in os.listdir('motivatingConstraints'):
        base, ext = os.path.splitext(filename)

        # Generate n copies of each file
        for i in range(1, n + 1):
            new_filename = f"{base}_{i}{ext}"
            new_filepath = os.path.join('motivatingConstraintsCopy_'+str(n), new_filename)

            # Copy the file contents
            with open(os.path.join('motivatingConstraints', filename), 'r') as f:
                contents = f.readlines()

            # Modify the package name in the first line of .rego files
            if ext == '.rego':
                new_package_name = f"package {base}_{i}"
                contents[0] = new_package_name + "\n"

            with open(new_filepath, 'w') as f:
                f.writelines(contents)

            print(f"Created {new_filename}")

# Example usage
generate_constraint_files(5)